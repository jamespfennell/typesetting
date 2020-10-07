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

// NextToken returns the next token in the input stream.
//
// The method retrieves one or more raw tokens and performs a number of processing steps, including:
// (1) Filtering out tokens of type catcode.Ignored.
// (2) Filtering out comments, and any whitespace after comments.
// (3) Combining multiple consecutive whitespace tokens into a single catcode.Space token,
// except when the consecutive whitespace tokens contain two newlines,
// in which case a "par" command token is returned.
// (4) Creating command tokens from tokens of type catcode.Escape and relevant tokens that follow.
//
// The method returns an error whenever the conditions described in NextRawToken are encountered.
//
// Because of these processing steps and error cases, this method will never return tokens with codes
// 0, 5, 9, 14 or 15. This is consistent with Exercise 7.3 of the TeXbook.
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
		t, err := tokenizer.NextRawToken()
		if err != nil {
			return t, err
		}
		switch t.catCode {
		case catcode.Ignored:
			continue
		case catcode.Invalid:
			tokenizer.err = errors.New(fmt.Sprintf("invalid character %s", t.Value()))
			return t, tokenizer.err
		case catcode.Escape:
			return tokenizer.readCommand()
		case catcode.Comment:
			for t.catCode != catcode.EndOfLine {
				t, err = tokenizer.NextRawToken()
				if err != nil {
					return t, err
				}
			}
			tokenizer.swallowNextWhitespace = true
			fallthrough
		case catcode.Space:
			fallthrough
		case catcode.EndOfLine:
			// We re-use the first token for the composite space token because we want to keep its metadata
			firstToken := t
			var b strings.Builder
			numEndOfLines := 0
			for t.catCode == catcode.Space || t.catCode == catcode.EndOfLine {
				b.WriteString(t.Value())
				if t.catCode == catcode.EndOfLine {
					numEndOfLines++
				}
				t, err = tokenizer.NextRawToken()
				if err != nil {
					return t, err
				}
			}
			_ = tokenizer.reader.UnreadRune()
			if numEndOfLines > 1 {
				return NewCommandToken("par"), nil
			}
			if tokenizer.swallowNextWhitespace {
				continue
			}
			firstToken.value = b.String()
			firstToken.catCode = catcode.Space
			return firstToken, nil
		default:
			return t, nil
		}
	}
}

// NextRawToken returns the next token in the Tokenizer as read directly from the input stream (hence "raw") and
// without doing any processing relating to comments or spacing or commands. This method should not be used in general
// and is only exposed for debugging purposes.
//
// An error is returned in the following 3 circumstances.
// (1) The end of the input stream has been reached, in which case the error will be io.EOF.
// (2) There is an error retrieving an element from the input stream.
// (3) The next element in the stream is not a valid UTF-8 character.
func (tokenizer *Tokenizer) NextRawToken() (Token, error) {
	r, _, err := tokenizer.reader.ReadRune()
	if err != nil {
		tokenizer.err = err
		// Potentially this is end of input error io.EOF
		return NewNullToken(), tokenizer.err
	}
	var c catcode.CatCode
	s := string(r)
	if r == unicode.ReplacementChar {
		tokenizer.err = errors.New("not a valid UTF-8 character")
		c = catcode.Invalid
	} else {
		c = tokenizer.catCodeMap.Get(s)
	}
	return NewCharacterToken(s, c), tokenizer.err
}

func (tokenizer *Tokenizer) readCommand() (Token, error) {
	t, err := tokenizer.NextRawToken()
	if err != nil {
		return t, err
	}
	var b strings.Builder
	b.WriteString(t.Value())
	for t.CatCode() == catcode.Letter {
		t, err = tokenizer.NextRawToken()
		if err != nil || t.CatCode() != catcode.Letter {
			_ = tokenizer.reader.UnreadRune()
			break
		}
		b.WriteString(t.Value())
	}
	return NewCommandToken(b.String()), nil
}
