package expansion

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

func Expand(ctx context.Context, list stream.TokenStream) stream.TokenStream {
	return &expandingStream{ctx: ctx, input: list}
}

type expandingStream struct {
	ctx       context.Context
	input     stream.TokenStream
	cmdOutput stream.TokenStream // TODO: make the type of this *expandingStream!
}

// TODO tests and then clean up
func (s *expandingStream) NextToken() (token.Token, error) {
	return s.nextOrPeek(stream.NextTokenOp)
}

func (s *expandingStream) PeekToken() (token.Token, error) {
	return s.nextOrPeek(stream.PeekTokenOp)
}

func (s *expandingStream) nextOrPeek(op func(stream.TokenStream) (token.Token, error)) (token.Token, error) {
	for {
		if s.cmdOutput != nil {
			t, err := op(s.cmdOutput)
			if err != nil || t != nil {
				return t, err
			}
			s.cmdOutput = nil
		}
		t, err := op(s.input)
		if err != nil || t == nil {
			return t, err
		}
		if !t.IsCommand() {
			return t, nil
		}
		cmd, ok := s.ctx.Registry.GetCommand(t.Value())
		if !ok {
			fmt.Println("Error: unknown command", t.Value())
			return t, nil
		}
		expansionCmd, ok := cmd.(CanonicalFunc)
		if !ok {
			fmt.Println("Skipping non-expansion command", t.Value())
			return t, nil
		}
		cmdOutput := expansionCmd(s.ctx, s.input)
		s.cmdOutput = Expand(s.ctx, cmdOutput)
		continue
	}
}
