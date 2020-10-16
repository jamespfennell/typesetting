package tex

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/command/library"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
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
	ctx.CatCodeMap = catcode.NewCatCodeMapWithTexDefaults()
	expansion.Register(&ctx.Registry, "input", library.Input)
	expansion.Register(&ctx.Registry, "string", library.String)
	expansion.Register(&ctx.Registry, "year", library.Year)
	return ctx
}

func runInternal(ctx *context.Context, filePath string) error {

	tokenList := input.NewTokenizerFromFilePath(ctx, filePath)
	expandedList := stream.NewStreamWithLog(expansion.Expand(ctx, tokenList), ctx.ExpansionLog)

	return stream.Consume(expandedList)
}
