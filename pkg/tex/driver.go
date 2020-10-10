package tex

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"os"
	"strings"
)

func Run(filePath string) {
	err := runInternal(filePath)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

func String(ctx context.Context, tokenStream stream.TokenStream) (stream.TokenStream, error) {
	t, err := tokenStream.NextToken()
	if err != nil {
		// TODO: we're converting a token error to a function error :/
		return nil, err
	}
	if t == nil {
		// TODO: don't return a null stream
		return nil, nil
	}
	if !t.IsCommand() {
		return stream.NewSingletonStream(t), nil
	}
	// TODO: we know the capacity, use it
	tokens := []token.Token{
		token.NewCharacterToken("\\", catcode.Other),
	}
	for _, c := range t.Value() {
		// TODO: space should have catcode space apparently.
		tokens = append(tokens, token.NewCharacterToken(string(c), catcode.Other))
	}
	return stream.NewSliceStream(tokens), nil
}

func Year(context.Context, stream.TokenStream) (stream.TokenStream, error) {
	// TODO: it won't always be 2020 :)
	return stream.NewSliceStream(
		[]token.Token{
			token.NewCharacterToken("2", catcode.Other),
			token.NewCharacterToken("0", catcode.Other),
			token.NewCharacterToken("2", catcode.Other),
			token.NewCharacterToken("0", catcode.Other),
		}), nil
}

func runInternal(filePath string) error {
	m1 := catcode.NewCatCodeMapWithTexDefaults() // TODO: return references to maps instead
	m2 := context.NewCommandMap()

	// tokenizationChannel := make(chan token.Token)
	// go outputTokenization(tokenizationChannel)
	ctx := context.Context{
		CatCodeMap: &m1,
		CommandMap: &m2,
		//TokenizerChannel: tokenizationChannel,
	}
	ctx.RegisterExpansionCommand("string", String)
	ctx.RegisterExpansionCommand("year", Year)
	// TODO: add functions

	tokenList, err := input.NewTokenizerFromFilePath(filePath, &m1)
	if err != nil {
		return err
	}
	expandedList := expansion.Expand(ctx, tokenList)

	for {
		t, err := expandedList.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return nil
		}
		fmt.Println(t)
	}
}

func outputTokenization(c <-chan token.Token) {
	for t := range c {
		var b strings.Builder
		if t.CatCode() < 0 {
			b.WriteString("cmd")
		} else {
			b.WriteString(fmt.Sprintf("%3d", t.CatCode()))
		}
		b.WriteString(" | ")
		if t.Value() == "\n" {
			b.WriteString("<newline>")
		} else {
			b.WriteString(t.Value())
		}
		fmt.Println(b.String())
	}
}
