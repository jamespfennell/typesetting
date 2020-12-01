// Package token contains the definitions of the Token and Stream types, which are fundamental in GoTex.
package token

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

// Token represents a single atom of input.
type Token interface {
	Value() string
	CatCode() catcode.CatCode
	IsCommand() bool
	Source() Source
	Description() string
}

// Source contains information on the origin on a token. It is used for printing helpful error messages when
// a user's TeX file has an error.
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
	// NextToken retrieves the next Token in the Stream and removes it from the Stream.
	NextToken() (Token, error)

	// PeekToken retrieves the next Token in the Stream without removing it from the Stream.
	PeekToken() (Token, error)
}

type ExpandingStream interface {
	Stream

	SourceStream() Stream
}

// Op abstractly represents one of the two stream operations: either NextToken, or PeekToken.
//
// This type exists because experience has shown that when creating new types of streams, the implementations of
// NextToken and PeekToken are often largely identical. Rather than duplicating code (with all the well-known caveats)
// Op enables creating a single implementation. See the definition of StreamWithOp for more information.
type Op struct {
	// Apply applies the operation (either next or peek) to the stream and returns the result. The return value
	// will be the same irrespective of the operation. However, the underlying state of the stream may be different
	// afterwards - if this is a peek operation, generally the state of the stream is not changed.
	Apply func(s Stream) (Token, error)

	// EnsureTokenConsumed ensures that the last token retrieved using Apply will be removed (or consumed) from
	// the stream. This function should generally be called after the state of the stream has been altered based on the
	// result of Apply. If this function is not called, and this is a peek operation, then the state change triggered
	// by the token may be erroneously repeated.
	EnsureTokenConsumed func(s Stream)

	// sealant prevents outside packages from implementing this type.
	sealant struct{}
}

var nextTokenOp = Op{
	Apply:               func(s Stream) (Token, error) { return s.NextToken() },
	EnsureTokenConsumed: func(_ Stream) {},
}

// NextTokenOp returns the Op representing NextToken.
func NextTokenOp() Op {
	return nextTokenOp
}

var peekTokenOp = Op{
	Apply:               func(s Stream) (Token, error) { return s.PeekToken() },
	EnsureTokenConsumed: func(s Stream) { _, _ = s.NextToken() },
}

// NextTokenOp returns the Op representing PeekToken.
func PeekTokenOp() Op {
	return peekTokenOp
}

// StreamWithOp is a token stream with an additional method for applying an Op to the stream.
// Often when creating a new stream, one only needs to really implement the PerformOp function abstractly in terms
// of the Op. One can then create a stream from the type by defining NextToken and PeekToken like so:
//
//	func (s *NewStreamType) NextToken() (token.Token, error) {
//		return s.PerformOp(token.NextTokenOp())
//	}
//
//	func (s *NewStreamType) PeekToken() (token.Token, error) {
//		return s.PerformOp(token.PeekTokenOp())
//	}
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

// NewCommandToken returns a token representing a TeX control sequence
func NewCommandToken(value string, source Source) Token {
	return characterToken{value: value, catCode: -1, source: source}
}

// NewCharacterToken returns a token representing a single non-control character
func NewCharacterToken(value string, code catcode.CatCode, source Source) Token {
	return characterToken{value: value, catCode: code, source: source}
}

// ErrorOrNil returns true if the token is nil or the error is non-nil
func ErrorOrNil(t Token, err error) bool {
	return err != nil || t == nil
}
