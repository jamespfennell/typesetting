package testutil

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"strings"
	"testing"
)

func RunExpansionTest(t *testing.T, ctx *context.Context, input, expectedOutput string) {
	startingStream := NewStream(ctx, input)
	expectedStream := NewStream(ctx, expectedOutput)
	actualStream := expansion.Expand(ctx, startingStream)

	CheckStreamEqual(t, expectedStream, actualStream)
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
	return input.NewTokenizer(ctx, strings.NewReader(s))
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
	ctx.CatCodeMap = catcode.NewCatCodeMapWithTexDefaults()
	return ctx
}
