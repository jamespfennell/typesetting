package expansion_test

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	. "github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"testing"
)

func TestExpand_BasicCase(t *testing.T) {
	ctx := context.NewContext()
	RegisterFunc(ctx, "funca", func() token.Stream {
		return testutil.NewSimpleStream("a1", "a2")
	})

	inputStream := testutil.NewSimpleStream("01", "funca", "02")
	expectedStream := testutil.NewSimpleStream("01", "a1", "a2", "02")
	actualStream := Expand(ctx, inputStream)

	testutil.CheckStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_InputProcessing(t *testing.T) {
	ctx := context.NewContext()
	RegisterFunc(ctx, "funca", func(c *context.Context, s token.Stream) token.Stream {
		t, _ := s.NextToken()
		return testutil.NewSimpleStream("a1", t.Value(), "a2")
	})

	inputStream := testutil.NewSimpleStream("01", "funca", "02")
	expectedStream := testutil.NewSimpleStream("01", "a1", "02", "a2")
	actualStream := Expand(ctx, inputStream)

	testutil.CheckStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_StackIsConsumed(t *testing.T) {
	ctx := context.NewContext()
	RegisterFunc(ctx, "funca", func() token.Stream {
		return testutil.NewSimpleStream("funcb", "a1")
	})
	RegisterFunc(ctx, "funcb", func(c *context.Context, s token.Stream) token.Stream {
		_, _ = s.NextToken()
		_, _ = s.NextToken()
		return testutil.NewSimpleStream("b1", "b2")
	})

	startingStream := testutil.NewSimpleStream("01", "funca", "02", "03")
	expectedStream := testutil.NewSimpleStream("01", "b1", "b2", "03")
	actualStream := Expand(ctx, startingStream)

	testutil.CheckStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_InputToExpansionFunctionNotExpanded(t *testing.T) {
	ctx := context.NewContext()
	RegisterFunc(ctx, "funca", func(c *context.Context, s token.Stream) token.Stream {
		_, _ = s.NextToken()
		return testutil.NewSimpleStream("a1", "a2")
	})
	RegisterFunc(ctx, "funcb", func() token.Stream {
		return testutil.NewSimpleStream("b1", "b2")
	})

	startingStream := testutil.NewSimpleStream("01", "funca", "funcb")
	expectedStream := testutil.NewSimpleStream("01", "a1", "a2")
	actualStream := Expand(ctx, startingStream)

	testutil.CheckStreamEqual(t, expectedStream, actualStream)
}

func TestExpand_NonExpandableCommandsPassThrough(t *testing.T) {
	ctx := context.NewContext()

	startingStream := testutil.NewSimpleStream("funca")
	expectedStream := testutil.NewSimpleStream("funca")
	actualStream := Expand(ctx, startingStream)

	testutil.CheckStreamEqual(t, expectedStream, actualStream)
}
