package library

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

func Input(ctx context.Context, s stream.TokenStream) stream.TokenStream {
	filePath, err := stream.ReadString(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	return input.NewTokenizerFromFilePath(filePath, ctx.CatCodeMap)
}
