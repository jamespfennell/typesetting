package conditional

import (
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"strconv"
	"testing"
)

func Test_IfTrue(t *testing.T) {
	paramsList := []struct {
		input  string
		output string
	}{
		{
			"\\iftrue abc\\fi",
			"abc",
		},
		{
			"a\\iftrue b\\fi c",
			"abc",
		},
		{
			"\\iftrue \\else abc\\fi",
			"",
		},
		{
			"\\iftrue\\else\\iftrue def\\fi abc\\fi",
			"",
		},
		{
			"\\iftrue\\iftrue\\else def\\fi\\else abc\\fi",
			"",
		},
		{
			"\\iffalse abc\\fi def",
			"def",
		},
		{
			"\\iffalse \\else abc\\fi",
			"abc",
		},
		{
			"a\\iffalse b\\else c\\fi d",
			"acd",
		},
	}
	for i, params := range paramsList {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			ctx := testutil.CreateTexContext()
			expansion.Register(ctx, "else", GetElse())
			expansion.Register(ctx, "fi", GetFi())
			expansion.Register(ctx, "iftrue", GetIfTrue())
			expansion.Register(ctx, "iffalse", GetIfFalse())

			testutil.RunExpansionTest(t, ctx, params.input, params.output)
		})
	}

}
