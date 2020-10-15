package logging

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"strings"
)

type LogEntry struct {
	t token.Token
	s string
	e error
}

func NewTokenizerLogPair() (LogSender, TokenizerLogReceiver) {
	c := make(chan LogEntry)
	finalizer := make(chan struct{})
	return LogSender{c: c, finalizer: finalizer}, TokenizerLogReceiver{c: c, finalizer: finalizer}
}

type LogSender struct {
	c chan<- LogEntry
	finalizer <-chan struct{}
}

func (sender LogSender) SendComment(comment string) {
	if sender.c != nil {
		sender.c <- LogEntry{s: comment}
	}
}

func (sender LogSender) SendToken(t token.Token, e error) {
	if sender.c != nil {
		sender.c <- LogEntry{t: t, e: e}
	}
}

func (sender LogSender) Close() {
	if sender.c != nil {
		close(sender.c)
	}
	if sender.finalizer != nil {
		<-sender.finalizer
	}
}

type TokenizerLogReceiver struct {
	c <-chan LogEntry
	finalizer chan<- struct{}
}

func (receiver TokenizerLogReceiver) Run() {
	fmt.Println("% GoTex tokenizer output")
	fmt.Println("%")
	fmt.Println("% This output is valid TeX and is equivalent to the input")
	fmt.Println("%")
	fmt.Println(fmt.Sprintf("%%%14s | value", "catcode"))
	for entry := range receiver.c {
		switch true {
		case entry.e != nil:
			fmt.Println(entry.e.Error())
		case entry.t != nil:
			var b strings.Builder
			if entry.t.CatCode() < 0 {
				b.WriteString("\\")
				b.WriteString(fmt.Sprintf("%-10s ", entry.t.Value() + "%"))
				b.WriteString("cmd")
			} else {
				if entry.t.CatCode() == catcode.Space {
					b.WriteString(" ")
				} else {
					b.WriteString(entry.t.Value())
				}
				b.WriteString(fmt.Sprintf("%-10s ", "%"))
				b.WriteString(fmt.Sprintf("%3d", entry.t.CatCode()))
			}
			b.WriteString(" | ")
			if entry.t.Value() == "\n" {
				b.WriteString("<newline>")
			} else {
				b.WriteString(entry.t.Value())
			}
			fmt.Println(b.String())
		}
	}
	receiver.finalizer <- struct{}{}
}