package commands

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization"
)

func Input(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	filePath, err := stream.ReadString(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	return tokenization.NewTokenizerFromFilePath(ctx, filePath)
}
