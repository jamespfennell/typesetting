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
	ctx.Tokenization.CatCodes = catcode.NewCatCodeMapWithTexDefaults()
	expansion.RegisterFunc(ctx, "input", library.Input)
	expansion.RegisterFunc(ctx, "string", library.String)
	expansion.RegisterFunc(ctx, "year", library.Year)
	return ctx
}

func runInternal(ctx *context.Context, filePath string) error {

	tokenList := input.NewTokenizerFromFilePath(ctx, filePath)
	expandedList := stream.NewStreamWithLog(expansion.Expand(ctx, tokenList), ctx.Expansion.Log)

	return stream.Consume(expandedList)
}
