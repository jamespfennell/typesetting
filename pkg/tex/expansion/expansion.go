package expansion

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

func Expand(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	stack := stream.NewStackStream()
	stack.Push(s)
	return &expansionStream{ctx: ctx, stack: stack}
}

type expansionStream struct {
	ctx   *context.Context
	stack *stream.StackStream
}

func (s *expansionStream) NextToken() (token.Token, error) {
	return s.PerformOp(stream.NextTokenOp)
}

func (s *expansionStream) PeekToken() (token.Token, error) {
	return s.PerformOp(stream.PeekTokenOp)
}

func (s *expansionStream) PerformOp(op stream.Op) (token.Token, error) {
	for {
		t, err := s.stack.PerformOp(op)
		if err != nil || t == nil || !t.IsCommand() {
			return t, err
		}
		cmd, ok := s.ctx.Registry.GetCommand(t.Value())
		if !ok {
			return t, NewUndefinedControlSequenceError(t)
		}
		expansionCmd, ok := cmd.(CanonicalFunc)
		if !ok {
			fmt.Println("Skipping non-expansion command", t.Value())
			return t, nil
		}
		s.stack.Push(expansionCmd(s.ctx, s.stack.Snapshot()))
	}
}

func NewUndefinedControlSequenceError(t token.Token) error {
	m := fmt.Sprintf("Undefined control sequence \\%s\n", t.Value())
	if t.Source() != nil {
		m += t.Source().String()
	}
	return errors.New(m)
}