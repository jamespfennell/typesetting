package logging

import (
	"github.com/jamespfennell/typesetting/pkg/tex/token"
)

type LogEntry struct {
	T token.Token
	S string
	E error
}

func NewLogPair() (LogSender, LogReceiver) {
	c := make(chan LogEntry)
	finalizer := make(chan struct{})
	return LogSender{c: c, finalizer: finalizer}, LogReceiver{c: c, finalizer: finalizer}
}

type LogSender struct {
	c         chan<- LogEntry
	finalizer <-chan struct{}
}

func (sender LogSender) SendComment(comment string) {
	if sender.c != nil {
		sender.c <- LogEntry{S: comment}
	}
}

func (sender LogSender) SendToken(t token.Token, e error) {
	if sender.c != nil {
		sender.c <- LogEntry{T: t, E: e}
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

type LogReceiver struct {
	c         <-chan LogEntry
	finalizer chan<- struct{}
}

func (receiver *LogReceiver) GetEntry() (LogEntry, bool) {
	entry, ok := <-receiver.c
	if !ok {
		receiver.finalizer <- struct{}{}
		return LogEntry{}, false
	}
	return entry, true
}
