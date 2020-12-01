package expansion

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

func Register(registry *context.Context, name string, cmd context.ExpansionCommand) {
	registry.Expansion.Commands.Set(name, cmd)
}

func RegisterFunc(registry *context.Context, name string, rawF interface{}) {
	registry.Expansion.Commands.Set(name, castFuncToExpansionCmd(rawF))
}

type Func000 func() []token.Token
type Func002 func() token.Stream
type Func010 func(s token.Stream) []token.Token
type Func111 func(ctx *context.Context, s token.Stream) ([]token.Token, error)
type Func112 func(ctx *context.Context, s token.Stream) token.Stream

func (f Func000) Invoke(ctx *context.Context, s token.Stream) token.Stream {
	return stream.NewSliceStream(f())
}

func (f Func002) Invoke(ctx *context.Context, s token.Stream) token.Stream {
	return f()
}

func (f Func010) Invoke(ctx *context.Context, s token.Stream) token.Stream {
	return stream.NewSliceStream(f(s))
}

func (f Func111) Invoke(ctx *context.Context, s token.Stream) token.Stream {
	slice, err := f(ctx, s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	return stream.NewSliceStream(slice)
}

func (f Func112) Invoke(ctx *context.Context, s token.Stream) token.Stream {
	return f(ctx, s)
}

func castFuncToExpansionCmd(rawF interface{}) context.ExpansionCommand {
	switch castF := rawF.(type) {
	case func() []token.Token:
		return Func000(castF)
	case func() token.Stream:
		return Func002(castF)
	case func(s token.Stream) []token.Token:
		return Func010(castF)
	case func(ctx *context.Context, s token.Stream) ([]token.Token, error):
		return Func111(castF)
	case func(ctx *context.Context, s token.Stream) token.Stream:
		return Func112(castF)
	}
	panic(
		fmt.Sprintf(
			"unable to convert the provided type to an expansion command.\n"+
				"Problematic type: %T\n"+
				"To resolve this, either:\n"+
				"  1) Change the type to be one of the built-in FuncXXX types in the expansion package.\n"+
				"  2) Make the type satisfy the context.ExpansionCommand interface.",
			rawF,
		),
	)
}
