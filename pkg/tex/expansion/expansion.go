package expansion

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"strings"
)

func Expand(ctx *context.Context, s token.Stream) token.ExpandingStream {
	stack := stream.NewStackStream()
	stack.Push(s)
	return &expansionStream{ctx: ctx, stack: stack}
}

// TODO: what is this about???
type loggingStream struct {
	*stream.StackStream
	l logging.LogSender
}

func (s loggingStream) NextToken() (token.Token, error) {
	t, err := s.StackStream.NextToken()
	s.l.SendToken(t, err)
	return t, err
}

type expansionStream struct {
	ctx   *context.Context
	stack *stream.StackStream
}

func (s *expansionStream) NextToken() (token.Token, error) {
	var t token.Token
	var err error
	for {
		t, err = s.stack.NextToken()
		if err != nil || t == nil || !t.IsCommand() {
			break
		}
		cmd, ok := s.ctx.Expansion.Commands.Get(t.Value())
		// This may be an execution command. Undefined control sequence errors are handled in the executor
		if !ok {
			break
		}
		s.stack.Push(cmd.Invoke(s.ctx, s.stack.Snapshot()))
	}
	s.ctx.Expansion.Log.SendToken(t, err)
	return t, err
}

func (s *expansionStream) PeekToken() (token.Token, error) {
	var t token.Token
	var err error
	for {
		t, err = s.stack.PeekToken()
		if err != nil || t == nil || !t.IsCommand() {
			break
		}
		cmd, ok := s.ctx.Expansion.Commands.Get(t.Value())
		// This may be an execution command. Undefined control sequence errors are handled in the executor
		if !ok {
			break
		}
		// Consume the token now that we're acting on it
		_, _ = s.stack.NextToken()
		s.stack.Push(cmd.Invoke(s.ctx, s.stack.Snapshot()))
	}
	return t, err
}

func (s *expansionStream) SourceStream() token.Stream {
	return loggingStream{s.stack, s.ctx.Expansion.Log}
}

// Writer writes the output of the expansion process to stdout.
func Writer(receiver logging.LogReceiver) {
	fmt.Println("% GoTex expansion output")
	fmt.Println("%")
	fmt.Println("% This output is valid TeX and is equivalent to the tokenization")
	fmt.Println("%")
	var b strings.Builder
	curLine := 0
	for {
		entry, ok := receiver.GetEntry()
		if !ok {
			return
		}
		switch true {
		case entry.E != nil:

			// fmt.Println(entry.E.Error())
		case entry.T != nil:
			if entry.T.Source() != nil {
				readerSource, ok := entry.T.Source().(tokenization.ReaderSource)
				if ok && readerSource.LineIndex != curLine {
					curLine = readerSource.LineIndex
					fmt.Println(b.String())
					b.Reset()
				}
			}
			if entry.T.CatCode() < 0 {
				b.WriteString("\\")
				b.WriteString(entry.T.Value())
				b.WriteString(" ")
			} else {
				if entry.T.CatCode() == catcode.Space {
					b.WriteString(" ")
				} else {
					if entry.T.Value() == "\\" {
						b.WriteString("\\")
					}
					b.WriteString(entry.T.Value())
				}
			}

		case entry.S != "":
			fmt.Println("% " + entry.S)
		}
	}
}
