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
		{  // Definition is parsed successfully
			"\\def\\A{abc}",
			"",
		},
		{  // Output is correct
			"\\def\\A{abc}\\A",
			"abc",
		},
		{  // Multiple outputs
			"\\def\\A{abc}\\A\\A",
			"abcabc",
		},
		{  // One parameter parse successfully
			"\\def\\A#1{a-#1-b}",
			"",
		},
		{  // One undelimited parameter with correct output - single token input
			"\\def\\A#1{a-#1-b}\\A1",
			"a-1-b",
		},
		{  // One undelimited parameter with correct output - multiple token input
			"\\def\\A#1{a-#1-b}\\A{123}",
			"a-123-b",
		},
		{  // Two undelimited parameters - single token inputs
			"\\def\\A#1#2{#2-#1}\\A56",
			"6-5",
		},
		{  // Two undelimited parameters - multiple token inputs
			"\\def\\A#1#2{#2-#1}\\A{abc}{xyz}",
			"xyz-abc",
		},
		{  // Token prefix to consume
			"\\def\\A fgh{567}\\A fghi",
			"567i",
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
