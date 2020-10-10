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
}

func NewSingletonStream(token token.Token) TokenStream {
	return &singletonList{token: token}
}

type singletonList struct {
	token token.Token
}

func (list *singletonList) NextToken() (token.Token, error) {
	token := list.token
	list.token = nil
	return token, nil
}

func NewSliceStream(tokens []token.Token) TokenStream {
	return &sliceBasedList{tokens: tokens}
}

type sliceBasedList struct {
	tokens []token.Token
}

func (list *sliceBasedList) NextToken() (token.Token, error) {
	if len(list.tokens) == 0 {
		return nil, nil
	}
	token := list.tokens[0]
	list.tokens = list.tokens[1:]
	return token, nil
}

func NewChainedList(lists ...TokenStream) TokenStream {
	return &chainedList{lists: lists}
}

type chainedList struct {
	lists []TokenStream
}

func (list *chainedList) NextToken() (token.Token, error) {
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

func NewObservedList(list TokenStream, observationChan chan<- token.Token) TokenStream {
	return &observedList{list: list, observationChan: observationChan}
}

type observedList struct {
	list            TokenStream
	observationChan chan<- token.Token
}

func (list *observedList) NextToken() (token.Token, error) {
	token, err := list.list.NextToken()
	if err != nil || token == nil {
		close(list.observationChan)
	} else {
		list.observationChan <- token
	}
	return token, err
}

func NewListWithCleanup(list TokenStream, cleanupFunc func()) TokenStream {
	return &listWithCleanup{list: list, cleanupFunc: cleanupFunc}
}

type listWithCleanup struct {
	list        TokenStream
	cleanupFunc func()
	cleanedUp   bool
}

func (list *listWithCleanup) NextToken() (token.Token, error) {
	if list.cleanedUp {
		return nil, nil
	}
	token, err := list.list.NextToken()
	if err != nil || token == nil {
		list.cleanupFunc()
		list.cleanedUp = true
	}
	return token, err
}
