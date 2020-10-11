package library

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

func String(ctx context.Context, tokenStream stream.TokenStream) ([]token.Token, error) {
	t, err := tokenStream.NextToken()
	if err != nil || t == nil {
		return nil, err
	}
	if !t.IsCommand() {
		return []token.Token{t}, nil
	}
	// TODO: we know the capacity, use it
	tokens := []token.Token{
		// TODO: there is a register for
		token.NewCharacterToken("\\", catcode.Other),
	}
	for _, c := range t.Value() {
		// TODO: space should have catcode space apparently.
		tokens = append(tokens, token.NewCharacterToken(string(c), catcode.Other))
	}
	return tokens, nil
}
