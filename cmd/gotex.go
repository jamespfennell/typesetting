package main

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex"
	"github.com/jamespfennell/typesetting/pkg/tex/input"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"os"
)

func main() {
	if len(os.Args) <= 2 {
		fmt.Println("A command and file path must be provided!")
		os.Exit(1)
	}
	command := os.Args[1]
	filePath := os.Args[2]
	ctx := tex.CreateTexContext()

	switch command {
	case "tokenize":
		sender, receiver := logging.NewLogPair()
		defer sender.Close()
		go input.TokenizerWriter(receiver)
		ctx.TokenizerLog = sender
	case "expand":
		sender, receiver := logging.NewLogPair()
		defer sender.Close()
		go input.TokenizerWriter(receiver)
		ctx.ExpansionLog = sender
	}

	tex.Run(ctx, filePath)
}
