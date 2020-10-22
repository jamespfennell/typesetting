package macro

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"testing"
)

func TestDef(t *testing.T) {
	paramsList := []struct {
		input  string
		output string
	}{
		{ // Definition is parsed successfully
			"\\def\\A{abc}",
			"",
		},
		{ // Output is correct
			"\\def\\A{abc}\\A",
			"abc",
		},
		{ // Multiple outputs
			"\\def\\A{abc}\\A\\A",
			"abcabc",
		},
		{ // One parameter parse successfully
			"\\def\\A#1{a-#1-b}",
			"",
		},
		{ // One undelimited parameter with correct output
			"\\def\\A#1{a-#1-b}\\A1",
			"a-1-b",
		},
		{ // One undelimited parameter with correct output - multiple token input
			"\\def\\A#1{a-#1-b}\\A{123}",
			"a-123-b",
		},
		{ // Two undelimited parameters
			"\\def\\A#1#2{#2-#1}\\A56",
			"6-5",
		},
		{ // Two undelimited parameters - multiple token inputs
			"\\def\\A#1#2{#2-#1}\\A{abc}{xyz}",
			"xyz-abc",
		},
		{ // Token prefix to consume
			"\\def\\A fgh{567}\\A fghi",
			"567i",
		},
		{ // One undelimited parameter with prefix
			"\\def\\A abc#1{y#1z}\\A abcdefg",
			"ydzefg",
		},
		{ // One undelimited parameter with prefix - multiple token input
			"\\def\\A abc#1{y#1z}\\A abcdefg",
			"ydzefg",
		},
		{ // One delimited parameter
			"\\def\\A #1xxx{y#1z}\\A abcxxx",
			"yabcz",
		},
		{ // One delimited empty parameter
			"\\def\\A #1xxx{y#1z}\\A xxx",
			"yz",
		},
		{ // One delimited parameter with scope
			"\\def\\A #1xxx{#1}\\A abc{123xxx}xxx",
			"abc{123xxx}",
		},
		{ // One delimited parameter with prefix
			"\\def\\A a#1c{x#1y}\\A abcdef",
			"xbydef",
		},
		{ // Two delimited parameters with prefix
			"\\def\\A a#1c#2e{x#2y#1z}\\A abcdef",
			"xdybzf",
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
