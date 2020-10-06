package tokenizer

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"io"
	"strings"
	"unicode"
)

type Token struct {
	value   string
	catCode catcode.CatCode
}

func (token Token) Value() string {
	return token.value
}

func (token Token) IsCommand() bool {
	return token.catCode == -1
}

func (token Token) IsNull() bool {
	return token.catCode == -2
}

func (token Token) CatCode() catcode.CatCode {
	return token.catCode
}

func NewCommandToken(value string) Token {
	return Token{value: value, catCode: -1}
}

func NewCharacterToken(value string, code catcode.CatCode) Token {
	return Token{value: value, catCode: code}
}

func NewNullToken() Token {
	return Token{catCode: -2}
}

type Tokenizer struct {
	reader                *bufio.Reader
	catCodeMap            *catcode.Map
	swallowNextWhitespace bool
	err                   error
}

func NewTokenizer(input io.Reader, catCodeMap *catcode.Map) *Tokenizer {
	return &Tokenizer{
		reader:     bufio.NewReader(input),
		catCodeMap: catCodeMap,
	}
}

func (tokenizer *Tokenizer) NextToken() (Token, error) {
	if tokenizer.err != nil {
		return NewNullToken(), tokenizer.err
	}
	token, err := tokenizer.nextTokenInternal()
	if err == nil {
		tokenizer.swallowNextWhitespace = token.IsCommand()
	}
	return token, err
}

func (tokenizer *Tokenizer) nextTokenInternal() (Token, error) {
	for {
		s, catCode, err := tokenizer.readRune()
		if err != nil {
			return NewNullToken(), err
		}
		switch catCode {
		case catcode.Escape:
			command, err := tokenizer.readCommand()
			if err != nil {
				return NewNullToken(), err
			}
			return NewCommandToken(command), nil
		case catcode.Comment:
			for catCode != catcode.EndOfLine {
				_, catCode, err = tokenizer.readRune()
				if err != nil {
					return NewNullToken(), err
				}
			}
			tokenizer.swallowNextWhitespace = true
			fallthrough
		case catcode.Space:
			fallthrough
		case catcode.EndOfLine:
			numEndOfLines := 0
			for catCode == catcode.Space || catCode == catcode.EndOfLine {
				if catCode == catcode.EndOfLine {
					numEndOfLines++
				}
				_, catCode, err = tokenizer.readRune()
				if err != nil {
					return NewNullToken(), err
				}
			}
			_ = tokenizer.reader.UnreadRune()
			if numEndOfLines > 1 {
				return NewCommandToken("par"), nil
			}
			if tokenizer.swallowNextWhitespace {
				continue
			}
			return NewCharacterToken(s, catcode.Space), nil
		default:
			return NewCharacterToken(s, catCode), nil
		}
	}
}

func (tokenizer *Tokenizer) readRune() (string, catcode.CatCode, error) {
	for {
		r, _, err := tokenizer.reader.ReadRune()
		if r == unicode.ReplacementChar {
			err = errors.New("not a valid UTF-8 character")
		}
		if err != nil {
			tokenizer.err = err
			// Potentially this is end of input error io.EOF
			return "", catcode.Invalid, tokenizer.err
		}
		s := string(r)
		c := tokenizer.catCodeMap.Get(s)
		if c == catcode.Ignored {
			continue
		}
		if c == catcode.Invalid {
			tokenizer.err = errors.New(fmt.Sprintf("invalid character %s", s))
			return s, c, tokenizer.err
		}
		return s, c, nil
	}
}

func (tokenizer *Tokenizer) readCommand() (string, error) {
	s, catCode, err := tokenizer.readRune()
	if err != nil {
		return "", err
	}
	var b strings.Builder
	b.WriteString(s)
	for catCode == catcode.Letter {
		s, catCode, err = tokenizer.readRune()
		if catCode != catcode.Letter || err != nil {
			_ = tokenizer.reader.UnreadRune()
		} else {
			b.WriteString(s)
		}
	}
	return b.String(), nil
}
