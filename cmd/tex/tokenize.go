package main

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"os"
	"strings"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("A file path must be provided!")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Could not open file:", os.Args[1])
		os.Exit(2)
	}
	m := catcode.NewCatCodeMapWithTexDefaults()
	t := input.NewTokenizer(f, &m, nil)
	for token, err := t.NextToken(); err == nil; token, err = t.NextToken() {
		var b strings.Builder
		if token.CatCode() < 0 {
			b.WriteString("cmd")
		} else {
			b.WriteString(fmt.Sprintf("%3d", token.CatCode()))
		}
		b.WriteString(" | ")
		if token.Value() == "\n" {
			b.WriteString("<newline>")
		} else {
			b.WriteString(token.Value())
		}
		fmt.Println(b.String())
	}
}
