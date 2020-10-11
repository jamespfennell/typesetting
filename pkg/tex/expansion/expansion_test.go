package expansion

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/command"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"testing"
)

// FuncA returns two characters and then FuncB
func FuncA() []token.Token {
	return []token.Token{
		token.NewCharacterToken("a1", catcode.Letter),
		token.NewCommandToken("funcb"),
		token.NewCharacterToken("a2", catcode.Letter),
	}
}

// FuncB consumes two characters and then returns two characters
func FuncB(s stream.TokenStream) []token.Token {
	fmt.Println(s)
	a, _ := s.NextToken()
	fmt.Println("Deleted", a)
	b, berr := s.NextToken()
	fmt.Println("Deleted", b, berr)
	return []token.Token{
		token.NewCharacterToken("b1", catcode.Letter),
		token.NewCharacterToken("b2", catcode.Letter),
	}
}

// FuncC returns three characters
func FuncC() []token.Token {
	return []token.Token{
		token.NewCharacterToken("c1", catcode.Letter),
		token.NewCharacterToken("c2", catcode.Letter),
		token.NewCharacterToken("c3", catcode.Letter),
	}
}

func TestExpand(t *testing.T) {
	m2 := command.NewRegistry()
	ctx := context.Context{
		Registry: m2,
	}
	Register(m2, "funca", FuncA)
	Register(m2, "funcb", FuncB)
	Register(m2, "funcc", FuncC)

	startingStream := stream.NewSliceStream([]token.Token{
		token.NewCommandToken("funca"),
		token.NewCharacterToken("0", catcode.Letter),
		token.NewCommandToken("funcc"),
	})

	expectedStream := stream.NewSliceStream([]token.Token{
		token.NewCharacterToken("a1", catcode.Letter),
		token.NewCharacterToken("b1", catcode.Letter),
		token.NewCharacterToken("b2", catcode.Letter),
		token.NewCharacterToken("c1", catcode.Letter),
		token.NewCharacterToken("c2", catcode.Letter),
		token.NewCharacterToken("c3", catcode.Letter),
	})
	actualStream := Expand(ctx, startingStream)

	if !streamEqual(t, expectedStream, actualStream) {
		t.Errorf("Streams not the same!")
	}

}

func streamEqual(t *testing.T, s1, s2 stream.TokenStream) (result bool) {
	result = true
	for {
		t1, err1 := s1.NextToken()
		t2, err2 := s2.NextToken()
		fmt.Println("Retrieved token", t2)
		if err1 != err2 || t1 != t2 {
			t.Errorf("Difference! (%v, %v) != (%v, %v)", t1, err1, t2, err2)
			result = false
		}
		if err1 != nil || t1 == nil {
			return
		}
	}
}
