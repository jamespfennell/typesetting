package stream

import (
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
)

func NextTokenOp(s token.Stream) (token.Token, error) {
	return s.NextToken()
}

func PeekTokenOp(s token.Stream) (token.Token, error) {
	return s.PeekToken()
}

func NewSliceStream(tokens []token.Token) token.Stream {
	return &sliceStream{tokens: tokens}
}

type sliceStream struct {
	tokens []token.Token
}

func (s *sliceStream) NextToken() (token.Token, error) {
	if len(s.tokens) == 0 {
		return nil, nil
	}
	t := s.tokens[0]
	s.tokens = s.tokens[1:]
	return t, nil
}

func (s *sliceStream) PeekToken() (token.Token, error) {
	if len(s.tokens) == 0 {
		return nil, nil
	}
	return s.tokens[0], nil
}

func NewErrorStream(e error) token.Stream {
	return errorStream{e: e}
}

type errorStream struct {
	e error
}

func (s errorStream) NextToken() (token.Token, error) {
	return nil, s.e
}

func (s errorStream) PeekToken() (token.Token, error) {
	return nil, s.e
}

func NewChainedStream(s ...token.Stream) token.Stream {
	return chainedStream{streams: s}
}

type chainedStream struct {
	streams []token.Stream
}

func (s chainedStream) NextToken() (token.Token, error) {
	return s.PerformOp(NextTokenOp)
}

func (s chainedStream) PeekToken() (token.Token, error) {
	return s.PerformOp(PeekTokenOp)
}

func (s chainedStream) PerformOp(op token.Op) (token.Token, error) {
	for {
		if len(s.streams) == 0 {
			return nil, nil
		}
		t, err := op(s.streams[0])
		if err != nil {
			return t, err
		}
		if t == nil {
			s.streams = s.streams[1:]
			continue
		}
		return t, nil
	}
}

type StackStream struct {
	stack []token.Stream
}

func NewStackStream() *StackStream {
	return &StackStream{}
}

func (s *StackStream) Snapshot() *StackStream {
	c := *s
	return &c
}

func (s *StackStream) Push(ts token.Stream) {
	s.stack = append(s.stack, ts)
}

func (s *StackStream) NextToken() (token.Token, error) {
	return s.PerformOp(NextTokenOp)
}

func (s *StackStream) PeekToken() (token.Token, error) {
	return s.PerformOp(PeekTokenOp)
}

func (s *StackStream) PerformOp(op token.Op) (token.Token, error) {
	for {
		if len(s.stack) == 0 {
			return nil, nil
		}
		t, err := op(s.stack[len(s.stack)-1])
		if err != nil || t != nil {
			return t, err
		}
		s.stack = s.stack[:len(s.stack)-1]
	}
}

func NewStreamWithCleanup(list token.Stream, cleanupFunc func()) token.Stream {
	return &streamWithCleanup{list: list, cleanupFunc: cleanupFunc}
}

type streamWithCleanup struct {
	list        token.Stream
	cleanupFunc func()
	cleanedUp   bool
}

func (s *streamWithCleanup) NextToken() (token.Token, error) {
	if s.cleanedUp {
		return nil, nil
	}
	t, err := s.list.NextToken()
	if err != nil || t == nil {
		s.cleanupFunc()
		s.cleanedUp = true
	}
	return t, err
}

func (s *streamWithCleanup) PeekToken() (token.Token, error) {
	if s.cleanedUp {
		return nil, nil
	}
	return s.list.PeekToken() // TODO: what if the stream is over?
}

func NewStreamWithLog(s token.ExpandingStream, sender logging.LogSender) token.ExpandingStream {
	return streamWithLog{s, sender}
}

type streamWithLog struct {
	token.ExpandingStream
	logging.LogSender
}

func (s streamWithLog) NextToken() (token.Token, error) {
	t, err := s.ExpandingStream.NextToken()
	s.LogSender.SendToken(t, err)
	return t, err
}
