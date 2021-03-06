package commands

import (
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization"
)

func Input(ctx *context.Context, s token.Stream) token.Stream {
	filePath, err := stream.ReadString(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	return tokenization.NewTokenizerFromFilePath(ctx, filePath)
}
