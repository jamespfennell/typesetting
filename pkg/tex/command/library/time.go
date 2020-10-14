package library

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"time"
)

func Year() []token.Token {
	year := string(rune(time.Now().Year()))
	tokens := make([]token.Token, len(year))
	for i, c := range year {
		tokens[i] = token.NewCharacterToken(string(c), catcode.Other, nil) // TODO source
	}
	return tokens
}
