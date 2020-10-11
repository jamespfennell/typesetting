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
func (stream *expandingStream) NextToken() (token.Token, error) {
	for {
		if stream.cmdOutput != nil {
			t, err := stream.cmdOutput.NextToken()
			if err != nil {
				return t, err
			}
			if t != nil {
				return t, nil
			}
			stream.cmdOutput = nil
		}
		t, err := stream.input.NextToken()
		if err != nil {
			return t, err
		}
		if t == nil {
			return nil, nil
		}
		if t.IsCommand() {
			cmd, ok := stream.ctx.Registry.GetCommand(t.Value())
			if ok {
				expansionCmd, ok := cmd.(CanonicalFunc)
				if !ok {
					continue
				}
				cmdOutput := expansionCmd(stream.ctx, stream.input)
				// TODO: probably not good enough to just expand, need to also
				//  do def stuff and typesetting
				//  or MAYBE NOT
				// Definitely not. The whole point of the context is that downstream information
				// gets back up
				// Note could also implement without recursion using a stack of streams.
				// The tricky part is ensuring commands on the stack get the right input
				// Would probably need a reverse chain
				// AKA a stack of streams
				stream.cmdOutput = Expand(stream.ctx, cmdOutput)
				continue
			}
			fmt.Println("Unknown command, passing for the moment", t.Value())
		}
		return t, nil
		// If it is a command token, place its on the cmdOutput
		// And then feed it input from the input
		// When the output has ended, delete it and start reading from the input stream again
		/*
			if stream.ctx.TokenizerChannel != nil {
				stream.ctx.TokenizerChannel <- t
			}
			return t, nil
		*/
	}
}
