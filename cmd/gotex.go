package main

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex"
	"os"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("A file path must be provided!")
		os.Exit(1)
	}
	filePath := os.Args[1]
	tex.Run(filePath)
}
