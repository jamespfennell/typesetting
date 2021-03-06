package execution

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

func Execute(ctx *context.Context, s token.ExpandingStream) error {
	return ExecuteWithControl(ctx, s, s.NextToken, NewUndefinedControlSequenceError, DefaultNonCommandHandler)
}

func ExecuteWithControl(
	ctx *context.Context,
	s token.ExpandingStream,
	nextToken func() (token.Token, error),
	undefinedCommandHandler func(token.Token) error,
	nonCommandHandler func(*context.Context, token.ExpandingStream, token.Token) error,
) error {
	for {
		t, err := nextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return nil
		}
		if t.IsCommand() {
			cmd, ok := ctx.Execution.Commands.Get(t.Value())
			if !ok {
				if err := undefinedCommandHandler(t); err != nil {
					return err
				}
				continue
			}
			err := cmd.Invoke(ctx, s)
			if err != nil {
				return err
			}
			continue
		}
		if err := nonCommandHandler(ctx, s, t); err != nil {
			return err
		}
	}
}

func DefaultNonCommandHandler(ctx *context.Context, s token.ExpandingStream, t token.Token) error {
	switch t.CatCode() {
	case catcode.Escape:
		return errors.New("unexpected escape token")
	case catcode.BeginGroup:
		ctx.BeginScope()
	case catcode.EndGroup:
		// TODO: may not be able to end the scope
		ctx.EndScope()
	}
	return nil
}

func Register(ctx *context.Context, name string, cmd context.ExecutionCommand) {
	ctx.Execution.Commands.Set(name, cmd)
}

type Func11 func(ctx *context.Context, s token.ExpandingStream) error

func (f Func11) Invoke(ctx *context.Context, s token.ExpandingStream) error {
	return f(ctx, s)
}

func RegisterFunc(ctx *context.Context, name string, f func(*context.Context, token.ExpandingStream) error) {
	Register(ctx, name, Func11(f))
}

func NewUndefinedControlSequenceError(t token.Token) error {
	m := fmt.Sprintf("Undefined control sequence \\%s\n", t.Value())
	if t.Source() != nil {
		m += t.Source().String()
	}
	return errors.New(m)
}
