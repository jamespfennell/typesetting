package tex

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/command"
	"github.com/jamespfennell/typesetting/pkg/tex/command/library"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
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


func runInternal(filePath string) error {
	m1 := catcode.NewCatCodeMapWithTexDefaults() // TODO: return references to maps instead
	m2 := command.NewRegistry()

	// tokenizationChannel := make(chan token.Token)
	// go outputTokenization(tokenizationChannel)
	ctx := context.Context{
		CatCodeMap: &m1,
		Registry: m2,
		//TokenizerChannel: tokenizationChannel,
	}
	expansion.Register(ctx.Registry, "input", library.Input)
	expansion.Register(ctx.Registry, "string", library.String)
	expansion.Register(ctx.Registry, "year", library.Year)

	tokenList:= input.NewTokenizerFromFilePath(filePath, &m1)
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
