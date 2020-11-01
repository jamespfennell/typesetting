package macro

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/execution"
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"strconv"
	"strings"
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
		{ // One undelimited parameter with correct performReplacement
			"\\def\\A#1{a-#1-b}\\A1",
			"a-1-b",
		},
		{ // One undelimited parameter with correct performReplacement, parameter multiple times
			"\\def\\A#1{#1 #1 #1}\\A1",
			"1 1 1",
		},
		{ // One undelimited parameter with correct performReplacement - multiple token tokenization
			"\\def\\A#1{a-#1-b}\\A{123}",
			"a-123-b",
		},
		{ // Two undelimited values
			"\\def\\A#1#2{#2-#1}\\A56",
			"6-5",
		},
		{ // Two undelimited values - multiple token inputs
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
		{ // One undelimited parameter with prefix - multiple token tokenization
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
		{ // Two delimited values with prefix
			`\def\A a#1c#2e{x#2y#1z}\A abcdef`,
			"xdybzf",
		},
		{ // Two delimited values with prefix
			`\def\A #1c{x#1y}\A {Hello}c`,
			"xHelloy",
		},
		{ // TeXBook exercise 20.1
			"\\def\\mustnt{I must not talk in class.}" +
				"\\def\\five{\\mustnt\\mustnt\\mustnt\\mustnt\\mustnt}" +
				"\\def\\twenty{\\five\\five\\five\\five}" +
				"\\def\\punishment{\\twenty\\twenty\\twenty\\twenty\\twenty}" +
				"\\punishment ",
			strings.Repeat("I must not talk in class.", 100),
		},
		/* TODO: scoping needed for this to pass
		{ // TeXBook exercise 20.2
			"\\def\\a{\\b}" +
				"\\def\\b{A\\def\\a{B\\def\\a{C\\def\\a{\\b}}}}" +
				"\\def\\puzzle{\\a\\a\\a\\a\\a}",
				"ABCAB",
		},
		*/
		{ // TeXBook exercise 20.3, part 1
			"\\def\\row#1{(#1_1,\\ldots,#1_n)}\\row{\\bf x}",
			"(\\bf x_1,\\ldots,\\bf x_n)",
		},
		{ // TeXBook exercise 20.3, part 2
			"\\def\\row#1{(#1_1,\\ldots,#1_n)}\\row{{\\bf x}}",
			"({\\bf x}_1,\\ldots,{\\bf x}_n)",
		},
		{ // TeXBook exercise 20.4, part 1
			`\def\mustnt#1#2{I must not #1 in #2.}` +
				`\def\five#1#2{\mustnt{#1}{#2}\mustnt{#1}{#2}\mustnt{#1}{#2}\mustnt{#1}{#2}\mustnt{#1}{#2}}` +
				`\def\twenty#1#2{\five{#1}{#2}\five{#1}{#2}\five{#1}{#2}\five{#1}{#2}}` +
				`\def\punishment#1#2{\twenty{#1}{#2}\twenty{#1}{#2}\twenty{#1}{#2}\twenty{#1}{#2}\twenty{#1}{#2}}` +
				`\punishment{run}{the halls}`,
			strings.Repeat("I must not run in the halls.", 100),
		},
		{ // TeXBook exercise 20.4, part 2
			"\\def\\mustnt{I must not \\doit\\ in \\thatplace.}" +
				"\\def\\five{\\mustnt\\mustnt\\mustnt\\mustnt\\mustnt}" +
				"\\def\\twenty{\\five\\five\\five\\five}" +
				"\\def\\punishment#1#2{\\def\\doit{#1}\\def\\thatplace{#2}\\twenty\\twenty\\twenty\\twenty\\twenty}" +
				"\\punishment{run}{the halls}",
			strings.Repeat("I must not run\\ in the halls.", 100),
		},
		{ // TeXBook exercise 20.5
			"\\def\\a#1{\\def\\b##1{##1#1}}" +
				"\\a!\\b{Hello}",
			"Hello!",
		},
		{ // TeXBook example below exercise 20.5
			"\\def\\a#1#{\\hbox to #1}\\a3pt{x}",
			"\\hbox to 3pt{x}",
		},
		{ // TeXBook exercise 20.6
			"\\def\\b#1{And #1, World!}\\def\\a#{\\b}\\a{Hello}",
			"And Hello, World!",
		},
	}

	for i, params := range paramsList {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := context.NewContext()
			ctx.Tokenization.CatCodes = catcode.NewCatCodeMapWithTexDefaults()
			execution.Register(ctx, "def", GetDef())

			testutil.RunExpansionTest(t, ctx, params.input, params.output)
		})
	}
}

func TestDef_TeXBookExercise20dot7(t *testing.T) {
	ctx := context.NewContext()
	ctx.Tokenization.CatCodes = catcode.NewCatCodeMapWithTexDefaults()
	ctx.Tokenization.CatCodes.Set("[", catcode.BeginGroup)
	ctx.Tokenization.CatCodes.Set("]", catcode.EndGroup)
	ctx.Tokenization.CatCodes.Set("!", catcode.Parameter)
	execution.Register(ctx, "def", GetDef())

	testutil.RunExpansionTest(t, ctx,
		"\\def\\!!1#2![{!#]#!!2}\\! x{[y]][z}",
		"{#]![y][z}",
	)
}
