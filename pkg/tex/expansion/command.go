package expansion

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/command"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

// CanonicalFunc is a command.Command invoked during the expansion process.
type CanonicalFunc func(ctx context.Context, s stream.TokenStream) stream.TokenStream

// Func represents a function or other type that can be canonicalized to expansion CanonicalFunc
type Func interface {
	Canonicalize() CanonicalFunc
}

func Register(registry command.Registry, name string, rawF interface{}) {
	registry.SetCommand(name, castToFunc(rawF).Canonicalize())
}

type Func000 func() []token.Token
type Func010 func(s stream.TokenStream) []token.Token
type Func111 func(ctx context.Context, s stream.TokenStream) ([]token.Token, error)
type Func112 func(ctx context.Context, s stream.TokenStream) stream.TokenStream

func (f Func000) Canonicalize() CanonicalFunc {
	return func(ctx context.Context, s stream.TokenStream) stream.TokenStream {
		return stream.NewSliceStream(f())
	}
}

func (f Func010) Canonicalize() CanonicalFunc {
	return func(ctx context.Context, s stream.TokenStream) stream.TokenStream {
		return stream.NewSliceStream(f(s))
	}
}

func (f Func111) Canonicalize() CanonicalFunc {
	return func(ctx context.Context, s stream.TokenStream) stream.TokenStream {
		slice, err := f(ctx, s)
		if err != nil {
			return stream.NewErrorStream(err)
		}
		return stream.NewSliceStream(slice)
	}
}

func (f Func112) Canonicalize() CanonicalFunc {
	return CanonicalFunc(f)
}

func castToFunc(rawF interface{}) Func {
	switch castF := rawF.(type) {
	case Func:
		return castF
	case func() []token.Token:
		return Func000(castF)
	case func(s stream.TokenStream) []token.Token:
		return Func010(castF)
	case func(ctx context.Context, s stream.TokenStream) ([]token.Token, error):
		return Func111(castF)
	case func(ctx context.Context, s stream.TokenStream) stream.TokenStream:
		return Func112(castF)
	}
	panic(
		fmt.Sprintf(
			"unable to convert the provided type to an expansion command.\n"+
				"Problematic type: %T\n"+
				"To resolve this, either:\n"+
				"  1) Change the type to be one of the built-in FuncXXX types in the expansion package.\n"+
				"  2) Make the type satisfy the Func interface in the expansion package.",
			rawF,
		),
	)
}

/*
func (registry *Registry) RegisterExpansionCommand(name string, function ExpansionFunc) {
	registry.scopedMap.Set(name, function.Canonicalize())
}

func (registry *Registry) RegisterCommand(name string, function interface{}) {

}
type ExpansionFunc000 func() token.Token

func (f ExpansionFunc000) Canonicalize() ExpansionCommand {
	return func(ctx context.Context, s stream.TokenStream) stream.TokenStream {
		return stream.NewSingletonStream(f())
	}
}

type Func110 func(ctx context.Context, tokenStream stream.TokenStream) token.Token
type Func111 func(ctx context.Context, tokenStream stream.TokenStream) (token.Token, error)


*/
