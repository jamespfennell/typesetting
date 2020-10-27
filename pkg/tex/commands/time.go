package commands

import (
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"strconv"
	"time"
)

func Year() []token.Token {
	year := strconv.Itoa(time.Now().Year())
	tokens := make([]token.Token, len(year))
	for i, c := range year {
		tokens[i] = token.NewCharacterToken(string(c), catcode.Other, nil) // TODO source
	}
	return tokens
}
