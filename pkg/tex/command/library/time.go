package library

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
)

func Year() []token.Token {
	// TODO: it won't always be 2020 :)
	return []token.Token{
			token.NewCharacterToken("2", catcode.Other),
			token.NewCharacterToken("0", catcode.Other),
			token.NewCharacterToken("2", catcode.Other),
			token.NewCharacterToken("0", catcode.Other),
		}
}
