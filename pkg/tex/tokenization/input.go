package tokenization

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"io"
	"os"
	"strings"
	"unicode"
)

type Tokenizer struct {
	reader                *Reader //bufio.Reader
	catCodeMap            *catcode.Map
	logger                *logging.LogSender
	buffer                token.Token
	swallowNextWhitespace bool
	err                   error
	inputOver             bool
}

func NewTokenizerFromFilePath(ctx *context.Context, filePath string) stream.TokenStream {
	f, err := os.Open(filePath)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	ctx.Tokenization.Log.SendComment("Reading file: " + filePath)
	return stream.NewStreamWithCleanup(
		NewTokenizer(ctx, f),
		func() { _ = f.Close() },
	)
}

func NewTokenizer(ctx *context.Context, input io.Reader) *Tokenizer {
	return &Tokenizer{
		reader:     NewReader(input),
		catCodeMap: &ctx.Tokenization.CatCodes,
		logger:     &ctx.Tokenization.Log,
	}
}

// NextToken returns the next token in the tokenization stream.
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
func (tokenizer *Tokenizer) NextToken() (token.Token, error) {
	if tokenizer.err != nil {
		return nil, tokenizer.err
	}
	if tokenizer.inputOver {
		return nil, nil
	}
	if tokenizer.buffer != nil {
		t := tokenizer.buffer
		tokenizer.buffer = nil
		return t, nil
	}
	t, err := tokenizer.nextTokenInternal()
	if err == nil && t != nil {
		tokenizer.swallowNextWhitespace = t.IsCommand()
	}
	if tokenizer.logger != nil {
		tokenizer.logger.SendToken(t, err)
	}
	return t, err
}

func (tokenizer *Tokenizer) PeekToken() (token.Token, error) {
	t, err := tokenizer.NextToken()
	tokenizer.buffer = t
	return t, err
}

func (tokenizer *Tokenizer) nextTokenInternal() (token.Token, error) {
	for {
		t, err := tokenizer.NextRawToken()
		if err != nil || t == nil {
			return t, err
		}
		switch t.CatCode() {
		case catcode.Ignored:
			continue
		case catcode.Invalid:
			tokenizer.err = errors.New(fmt.Sprintf("invalid character %s", t.Value()))
			return t, tokenizer.err
		case catcode.Escape:
			return tokenizer.readCommand()
		case catcode.Comment:
			for t.CatCode() != catcode.EndOfLine {
				t, err = tokenizer.NextRawToken()
				if err != nil || t == nil {
					return t, err
				}
			}
			tokenizer.swallowNextWhitespace = true
			fallthrough
		case catcode.Space:
			fallthrough
		case catcode.EndOfLine:
			var b strings.Builder
			numEndOfLines := 0
			var source token.Source
			for t.CatCode() == catcode.Space || t.CatCode() == catcode.EndOfLine {
				source = t.Source()
				b.WriteString(t.Value())
				if t.CatCode() == catcode.EndOfLine {
					numEndOfLines++
				}
				t, err = tokenizer.NextRawToken()
				if err != nil || t == nil {
					return t, err
				}
			}
			_ = tokenizer.reader.UnreadRune()
			if numEndOfLines > 1 {
				return token.NewCommandToken("par", source), nil
			}
			if tokenizer.swallowNextWhitespace {
				continue
			}
			return token.NewCharacterToken(b.String(), catcode.Space, source), nil
		default:
			return t, nil
		}
	}
}

// NextRawToken returns the next token in the Tokenizer as read directly from the tokenization stream (hence "raw") and
// without doing any processing relating to comments or spacing or commands. This method should not be used in general
// and is only exposed for debugging purposes.
//
// An error is returned in the following 3 circumstances.
// (1) The end of the tokenization stream has been reached, in which case the error will be io.EOF.
// (2) There is an error retrieving an element from the tokenization stream.
// (3) The next element in the stream is not a valid UTF-8 character.
func (tokenizer *Tokenizer) NextRawToken() (token.Token, error) {
	r, _, err := tokenizer.reader.ReadRune()
	if err != nil {
		if err == io.EOF {
			tokenizer.inputOver = true
			return nil, nil
		}
		tokenizer.err = err
		return nil, tokenizer.err
	}
	var c catcode.CatCode
	s := string(r)
	if r == unicode.ReplacementChar {
		tokenizer.err = errors.New("not a valid UTF-8 character")
		c = catcode.Invalid
	} else {
		c = tokenizer.catCodeMap.Get(s)
	}
	lineIndex, runeIndex := tokenizer.reader.Coordinates()
	source := ReaderSource{
		line:              tokenizer.reader.stringLine,
		reader:            tokenizer.reader,
		LineIndex:         lineIndex,
		startingRuneIndex: runeIndex,
		endingRuneIndex:   runeIndex,
	}
	return token.NewCharacterToken(s, c, source), tokenizer.err
}

func (tokenizer *Tokenizer) readCommand() (token.Token, error) {
	lineIndex, runeIndex := tokenizer.reader.Coordinates()
	t, err := tokenizer.NextRawToken()
	if err != nil || t == nil {
		return t, err
	}
	var b strings.Builder
	b.WriteString(t.Value())
	for t.CatCode() == catcode.Letter {
		t, err = tokenizer.NextRawToken()
		if err != nil || t == nil || t.CatCode() != catcode.Letter {
			_ = tokenizer.reader.UnreadRune()
			break
		}
		b.WriteString(t.Value())
	}
	value := b.String()
	source := ReaderSource{
		line:              tokenizer.reader.stringLine,
		reader:            tokenizer.reader,
		LineIndex:         lineIndex,
		startingRuneIndex: runeIndex,
		endingRuneIndex:   runeIndex + len(value),
	}
	return token.NewCommandToken(b.String(), source), nil
}

type ReaderSource struct {
	line              string
	reader            *Reader // TODO: remove this field
	startingRuneIndex int
	endingRuneIndex   int
	LineIndex         int
}

func (source ReaderSource) String() string {
	var b strings.Builder
	b.WriteString(
		fmt.Sprintf(
			"In file \"tokenization.tex\", line %d, char %d:\n",
			source.LineIndex+1,
			source.startingRuneIndex+1,
		),
	)
	b.WriteString(">  ")
	b.WriteString(source.line)
	b.WriteString("\n")
	b.WriteString(strings.Repeat(" ", source.startingRuneIndex+3))
	b.WriteString(strings.Repeat("^", source.endingRuneIndex-source.startingRuneIndex+1))
	return b.String()
}

// TokenizerWriter writes the output of the tokenizer to stdout.
func TokenizerWriter(receiver logging.LogReceiver) {
	fmt.Println("% GoTex tokenizer output")
	fmt.Println("%")
	fmt.Println("% This output is valid TeX and is equivalent to the tokenization")
	fmt.Println("%")
	fmt.Println(fmt.Sprintf("%%%14s | value", "catcode"))
	for {
		entry, ok := receiver.GetEntry()
		if !ok {
			return
		}
		switch true {
		case entry.E != nil:
			fmt.Println(entry.E.Error())
		case entry.T != nil:
			var b strings.Builder
			if entry.T.CatCode() < 0 {
				b.WriteString("\\")
				b.WriteString(fmt.Sprintf("%-10s ", entry.T.Value()+"%"))
				b.WriteString("cmd")
			} else {
				if entry.T.CatCode() == catcode.Space {
					b.WriteString(" ")
				} else {
					b.WriteString(entry.T.Value())
				}
				b.WriteString(fmt.Sprintf("%-10s ", "%"))
				b.WriteString(fmt.Sprintf("%3d", entry.T.CatCode()))
			}
			b.WriteString(" | ")
			if entry.T.Value() == "\n" {
				b.WriteString("<newline>")
			} else {
				b.WriteString(entry.T.Value())
			}
			fmt.Println(b.String())
		case entry.S != "":
			fmt.Println("%               | " + entry.S)
		}
	}
}
