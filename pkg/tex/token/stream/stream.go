package stream

import "github.com/jamespfennell/typesetting/pkg/tex/token"

// TokenStream represents a token list, a fundamental data type in TeX.
// A token list is an ordered collection of Token types which are retrieved on demand.
//
// The name 'token list' originates in the TeX82 implementation.
// Conceptually it's better to think of a token list as a stream instead of a list:
// in many cases the contents of the list is not predetermined when the list is created.
type TokenStream interface {
	// NextToken retrieves the next Token in the TokenStream.
	NextToken() (token.Token, error)

	PeekToken() (token.Token, error)

}

func NextTokenOp(s TokenStream) (token.Token, error) {
	return s.NextToken()
}

func PeekTokenOp(s TokenStream) (token.Token, error) {
	return s.PeekToken()
}

func NewSingletonStream(token token.Token) TokenStream {
	return &singletonStream{token: token}
}

type singletonStream struct {
	token token.Token
}

func (s *singletonStream) NextToken() (token.Token, error) {
	t := s.token
	s.token = nil
	return t, nil
}

func (s *singletonStream) PeekToken() (token.Token, error) {
	return s.token, nil
}

func NewSliceStream(tokens []token.Token) TokenStream {
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

func NewErrorStream(e error) TokenStream {
	return errorStream{e : e }
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

/*
func NewChainedStream(lists ...TokenStream) TokenStream {
	return &chainedStream{lists: lists}
}

type chainedStream struct {
	lists []TokenStream
}

func (list *chainedStream) NextToken() (token.Token, error) {
	for len(list.lists) > 0 {
		t, err := list.lists[0].NextToken()
		if err != nil {
			return t, err
		}
		if t == nil {
			list.lists = list.lists[1:]
			continue
		}
		return t, err
	}
	return nil, nil
}
 */


func NewStreamWithCleanup(list TokenStream, cleanupFunc func()) TokenStream {
	return &streamWithCleanup{list: list, cleanupFunc: cleanupFunc}
}

type streamWithCleanup struct {
	list        TokenStream
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
	return s.list.PeekToken()
}