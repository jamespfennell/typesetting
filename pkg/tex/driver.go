package tex

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/commands"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
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
	expansion.RegisterFunc(ctx, "tokenization", commands.Input)
	expansion.RegisterFunc(ctx, "string", commands.String)
	expansion.RegisterFunc(ctx, "year", commands.Year)
	return ctx
}

func runInternal(ctx *context.Context, filePath string) error {

	tokenList := tokenization.NewTokenizerFromFilePath(ctx, filePath)
	expandedList := stream.NewStreamWithLog(expansion.Expand(ctx, tokenList), ctx.Expansion.Log)

	return stream.Consume(expandedList)
}
