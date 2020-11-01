package testutil

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/execution"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"strings"
	"testing"
)

func RunExpansionTest(t *testing.T, ctx *context.Context, input, expectedOutput string) {
	startingStream := NewStream(ctx, input)
	expectedStream := NewStream(ctx, expectedOutput)

	var outputTokens []token.Token
	err := execution.ExecuteWithControl(
		ctx,
		expansion.Expand(ctx, startingStream),
		func(t token.Token) error {
			outputTokens = append(outputTokens, t)
			return nil
		},
		func(ctx *context.Context, s stream.ExpandingStream, t token.Token) error {
			switch t.CatCode() {
			case catcode.BeginGroup:
				ctx.BeginScope()
			case catcode.EndGroup:
				ctx.EndScope()
			}
			outputTokens = append(outputTokens, t)
			return nil
		},
	)
	if err != nil {
		t.Fatalf(err.Error())
	}
	actualStream := stream.NewSliceStream(outputTokens)
	CheckStreamEqual(t, expectedStream, actualStream)
}

// TODO: dedeplicate code or maybe not
func RunExpansionErrorTest(t *testing.T, ctx *context.Context, input string) error {
	startingStream := NewStream(ctx, input)
	err := execution.ExecuteWithControl(
		ctx,
		expansion.Expand(ctx, startingStream),
		func(t token.Token) error {
			return nil
		},
		func(ctx *context.Context, s stream.ExpandingStream, t token.Token) error {
			switch t.CatCode() {
			case catcode.BeginGroup:
				ctx.BeginScope()
			case catcode.EndGroup:
				ctx.EndScope()
			}
			return nil
		},
	)
	if err == nil {
		t.Errorf("Expected error, recieved none")
	}
	return err
}

func NewSimpleStream(values ...string) stream.TokenStream {
	var tokens []token.Token
	for _, value := range values {
		if len(value) >= 4 && value[:4] == "func" {
			tokens = append(tokens, token.NewCommandToken(value, nil))
		} else {
			tokens = append(tokens, token.NewCharacterToken(value, catcode.Letter, nil))
		}
	}
	return stream.NewSliceStream(tokens)
}

func NewStream(ctx *context.Context, s string) stream.TokenStream {
	return tokenization.NewTokenizer(ctx, strings.NewReader(s))
}

func CheckStreamEqual(t *testing.T, s1, s2 stream.TokenStream) (result bool) {
	result = true
	var v1, v2 string
	for {
		t1, err1 := s1.NextToken()
		t2, err2 := s2.NextToken()
		if err1 != err2 {
			t.Errorf("Error difference: %v != %v", err1, err2)
			result = false
		}
		if !CheckTokenEqual(t, t1, t2) {
			result = false
		}
		if t1 != nil {
			if t1.IsCommand() {
				v1 += `\` + t1.Value() + ` `
			} else {
				v1 += t1.Value()
			}
		}
		if t2 != nil {
			if t2.IsCommand() {
				v2 += `\` + t2.Value() + ` `
			} else {
				v2 += t2.Value()
			}
		}
		if err1 != nil || (t1 == nil && t2 == nil) {
			break
		}
	}
	if !result {
		t.Errorf("Full streams: \n%s\n%s", v1, v2)
	}
	return
}

func CheckTokenEqual(t *testing.T, t1, t2 token.Token) (result bool) {
	result = true
	switch true {
	case t1 == nil && t2 == nil:
		result = true
	case t1 == nil && t2 != nil:
		result = false
	case t1 != nil && t2 == nil:
		result = false
	case t1.Value() != t2.Value():
		result = false
	case t1.CatCode() != t2.CatCode():
		result = false
	}
	if !result {
		t.Errorf("Tokens not equal! %v != %v", t1, t2)
	}
	return
}

func CreateTexContext() *context.Context {
	ctx := context.NewContext()
	ctx.Tokenization.CatCodes = catcode.NewCatCodeMapWithTexDefaults()
	return ctx
}
