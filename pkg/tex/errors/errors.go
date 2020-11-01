package errors

import (
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"strings"
)

type UnexpectedTokenError struct {
	t        token.Token
	while    string // TODO: rename context and possibly with interface for advanced contexts
	expected string
	received string
}

func (err UnexpectedTokenError) Error() string {
	var b strings.Builder
	b.WriteString("unexpected token while ")
	b.WriteString(err.while)
	if err.t.Source() != nil {
		b.WriteString("\n")
		b.WriteString(err.t.Source().String())
	}
	b.WriteString("\n")
	b.WriteString("Expected: ")
	b.WriteString(err.expected)
	b.WriteString("\n")
	b.WriteString("Received: ")
	b.WriteString(err.received)
	return b.String()
}

func NewUnexpectedTokenError(t token.Token, expected, received, context string) UnexpectedTokenError {
	return UnexpectedTokenError{t, context, expected, received}
}

type UnexpectedEndOfInputError struct {
	while string
}

func (err UnexpectedEndOfInputError) Error() string {
	// TODO: this error message needs work - it would be ideal to print the last line of input if possible
	//  and also where the input stream started consuming
	return "unexpected end of input while " + err.while
}

func NewUnexpectedEndOfInputError(while string) UnexpectedEndOfInputError {
	return UnexpectedEndOfInputError{while}
}
