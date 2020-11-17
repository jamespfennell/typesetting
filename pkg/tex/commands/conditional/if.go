package conditional

import (
	"errors"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

type Condition func(ctx *context.Context, s stream.TokenStream) (bool, error)

type ifCmd struct {
	c Condition
}

func NewIfCommand(c Condition) context.ExpansionCommand {
	return ifCmd{c: c}
}

// GetIfTrue returns the \iftrue command, which always evaluates to true
func GetIfTrue() context.ExpansionCommand {
	return NewIfCommand(func(*context.Context, stream.TokenStream) (bool, error) { return true, nil })
}

// GetIfFalse returns the \iffalse command, which always evaluates to false
func GetIfFalse() context.ExpansionCommand {
	return NewIfCommand(func(*context.Context, stream.TokenStream) (bool, error) { return false, nil })
}

func (cmd ifCmd) Invoke(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	result, err := cmd.c(ctx, s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	if result {
		return &trueBranch{s: s, ctx: ctx}
	}
	return &falseBranch{s: s, ctx: ctx}
}

type elseCmd struct{}

func GetElse() context.ExpansionCommand {
	return elseCmd{}
}

func (cmd elseCmd) Invoke(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	return stream.NewErrorStream(errors.New("unexpected else command"))
}

type fiCmd struct{}

func GetFi() context.ExpansionCommand {
	return fiCmd{}
}

func (cmd fiCmd) Invoke(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	return stream.NewErrorStream(errors.New("unexpected fi command"))
}

func IsIfCommand(command context.ExpansionCommand) bool {
	_, ok := (command).(ifCmd)
	return ok
}

func IsElseCommand(cmd context.ExpansionCommand) bool {
	_, ok := (cmd).(elseCmd)
	return ok
}

func IsFiCommand(cmd context.ExpansionCommand) bool {
	_, ok := (cmd).(fiCmd)
	return ok
}
