package expansion

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"strings"
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
			return t, nil
			// return t, NewUndefinedControlSequenceError(t)
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

// Writer writes the output of the expansion process to stdout.
func Writer(receiver logging.LogReceiver) {
	fmt.Println("% GoTex expansion output")
	fmt.Println("%")
	fmt.Println("% This output is valid TeX and is equivalent to the input")
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
			fmt.Println(entry.E.Error())
		case entry.T != nil:
			if entry.T.Source() != nil {
				readerSource, ok := entry.T.Source().(input.ReaderSource)
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
