package expansion

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"testing"
)

func TestExpand_BasicCase(t *testing.T) {
	ctx := context.NewContext()
	Register(&ctx.Registry, "funca", func() stream.TokenStream {
		return newStream("a1", "a2")
	})

	inputStream := newStream("01", "funca", "02")
	expectedStream := newStream("01", "a1", "a2", "02")
	actualStream := Expand(ctx, inputStream)

	checkStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_InputProcessing(t *testing.T) {
	ctx := context.NewContext()
	Register(&ctx.Registry, "funca", func(c *context.Context, s stream.TokenStream) stream.TokenStream {
		t, _ := s.NextToken()
		return newStream("a1", t.Value(), "a2")
	})

	inputStream := newStream("01", "funca", "02")
	expectedStream := newStream("01", "a1", "02", "a2")
	actualStream := Expand(ctx, inputStream)

	checkStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_StackIsConsumed(t *testing.T) {
	ctx := context.NewContext()
	Register(&ctx.Registry, "funca", func() stream.TokenStream {
		return newStream("funcb", "a1")
	})
	Register(&ctx.Registry, "funcb", func(c *context.Context, s stream.TokenStream) stream.TokenStream {
		_, _ = s.NextToken()
		_, _ = s.NextToken()
		return newStream("b1", "b2")
	})

	startingStream := newStream("01", "funca", "02", "03")
	expectedStream := newStream("01", "b1", "b2", "03")
	actualStream := Expand(ctx, startingStream)

	checkStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_InputToExpansionFunctionNotExpanded(t *testing.T) {
	ctx := context.NewContext()
	Register(&ctx.Registry, "funca", func(c *context.Context, s stream.TokenStream) stream.TokenStream {
		_, _ = s.NextToken()
		return newStream("a1", "a2")
	})
	Register(&ctx.Registry, "funcb", func() stream.TokenStream {
		return newStream("b1", "b2")
	})

	startingStream := newStream("01", "funca", "funcb")
	expectedStream := newStream("01", "a1", "a2")
	actualStream := Expand(ctx, startingStream)

	checkStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_NonExpandableCommandsPassThrough(t *testing.T) {
	ctx := context.NewContext()
	ctx.Registry.SetCommand("funca", "dummy function value")

	startingStream := newStream("funca")
	expectedStream := newStream("funca")
	actualStream := Expand(ctx, startingStream)

	checkStreamEqual(t, expectedStream, actualStream)
}

func newStream(values ...string) stream.TokenStream {
	var tokens []token.Token
	for _, value := range values {
		if len(value) >= 4 && value[:4] == "func" {
			tokens = append(tokens, token.NewCommandToken(value))
		} else {
			tokens = append(tokens, token.NewCharacterToken(value, catcode.Letter))
		}
	}
	return stream.NewSliceStream(tokens)
}

func checkStreamEqual(t *testing.T, s1, s2 stream.TokenStream) (result bool) {
	result = true
	for {
		t1, err1 := s1.NextToken()
		t2, err2 := s2.NextToken()
		if err1 != err2 || t1 != t2 {
			t.Errorf("Difference! (%v, %v) != (%v, %v)", t1, err1, t2, err2)
			result = false
		}
		if err1 != nil || t1 == nil {
			return
		}
	}
}
