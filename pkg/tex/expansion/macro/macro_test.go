package macro

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"testing"
)

func TestDef(t *testing.T) {
	paramsList := []struct{
		input string
		output string
	}{
		{
			"\\def\\A{abc}",
			"",
		},
		{
			"\\def\\A{abc}\\A",
			"abc",
		},
		{
			"\\def\\A{abc}\\A\\A",
			"abcabc",
		},
		{
			"\\def\\A#1{a-#1-b}",
			"",
		},
		{
			"\\def\\A#1{a-#1-b}\\A1",
			"a-1-b",
		},
	}

	for _, params := range paramsList {
		t.Run(params.input, func(t *testing.T) {
			ctx := context.NewContext()
			ctx.CatCodeMap = catcode.NewCatCodeMapWithTexDefaults()
			expansion.Register(&ctx.Registry, "def", Def)

			startingStream := testutil.NewStream(ctx, params.input)
			expectedStream := testutil.NewStream(ctx, params.output)
			actualStream := expansion.Expand(ctx, startingStream)

			testutil.CheckStreamEqual(t, expectedStream, actualStream)
		})
	}
}
