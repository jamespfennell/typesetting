package tex

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/commands"
	"github.com/jamespfennell/typesetting/pkg/tex/commands/conditional"
	"github.com/jamespfennell/typesetting/pkg/tex/commands/macro"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/execution"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"os"
)

func Run(ctx *context.Context, filePath string) {
	err := runInternal(ctx, filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func CreateTexContext() *context.Context {
	ctx := context.NewContext()
	ctx.Tokenization.CatCodes = catcode.NewCatCodeMapWithTexDefaults()
	expansion.RegisterFunc(ctx, "input", commands.Input)
	expansion.RegisterFunc(ctx, "string", commands.String)
	expansion.RegisterFunc(ctx, "year", commands.Year)
	expansion.Register(ctx, "else", conditional.GetElse())
	expansion.Register(ctx, "fi", conditional.GetFi())
	expansion.Register(ctx, "iftrue", conditional.GetIfTrue())
	expansion.Register(ctx, "iffalse", conditional.GetIfFalse())

	execution.Register(ctx, "def", macro.GetDef())
	execution.RegisterFunc(ctx, "par", func(*context.Context, stream.ExpandingStream) error { return nil })
	execution.RegisterFunc(ctx, "relax", func(*context.Context, stream.ExpandingStream) error { return nil })
	return ctx
}

func runInternal(ctx *context.Context, filePath string) error {

	tokenList := tokenization.NewTokenizerFromFilePath(ctx, filePath)
	expandedList := expansion.Expand(ctx, tokenList)

	return execution.Execute(ctx, expandedList)
	// return stream.Consume(expandedList)
}
