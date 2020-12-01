package token

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

type Token interface {
	Value() string
	CatCode() catcode.CatCode
	IsCommand() bool
	Source() Source
	Description() string
}

type Source interface {
	String() string
}

// Stream represents a token list, a fundamental data type in TeX.
// A token list is an ordered collection of Token types which are retrieved on demand.
//
// The name 'token list' originates in the TeX82 implementation.
// Conceptually it's better to think of a token list as a stream instead of a list:
// in many cases the contents of the list is not predetermined when the list is created.
type Stream interface {
	// NextToken retrieves the next Token in the Stream.
	NextToken() (Token, error)

	PeekToken() (Token, error)
}

type ExpandingStream interface {
	Stream

	SourceStream() Stream
}

// TODO: need a way to consume in the peeking case.
//  Perhaps Op should be a struct containing the stream, and then have a Get and EnsureConsumed method
type Op func(s Stream) (Token, error)

type Op2 interface {
	Get(s Stream) (Token, error)
	EnsureTokenConsumed(s Stream)
	sealant()
}

type NextTokenOp struct{}

func (_ NextTokenOp) Get(s Stream) (Token, error) {
	return s.NextToken()
}

func (_ NextTokenOp) EnsureTokenConsumed(s Stream){}

func (_ NextTokenOp) sealant() {}
func aextTokenOp() Op2 {
	return NextTokenOp{}
}

type OpNew interface {
	Get(stream Stream) (Token, error)
	EnsureTokenConsumed(stream Stream)
}

type StreamWithOp interface {
	Stream

	PerformOp(Op) (Token, error)
}

type characterToken struct {
	value   string
	catCode catcode.CatCode
	source  Source
}

func (token characterToken) Value() string {
	return token.value
}

func (token characterToken) IsCommand() bool {
	return token.catCode == -1
}

func (token characterToken) CatCode() catcode.CatCode {
	return token.catCode
}

func (token characterToken) Source() Source {
	return token.source
}
func (token characterToken) Description() string {
	return fmt.Sprintf(
		"token with value %q and type %s (catcode = %d)",
		token.Value(), token.CatCode().String(), token.CatCode())
}

func (token characterToken) String() string {
	if token.CatCode() == -1 {
		return fmt.Sprintf("[cmd: %s]", token.Value())
	}
	return fmt.Sprintf("[val: %s; cc: %d]", token.Value(), token.CatCode())
}

func NewCommandToken(value string, source Source) Token {
	return characterToken{value: value, catCode: -1, source: source}
}

func NewCharacterToken(value string, code catcode.CatCode, source Source) Token {
	return characterToken{value: value, catCode: code, source: source}
}

func ErrorOrNil(t Token, err error) bool {
	return err != nil || t == nil
}
