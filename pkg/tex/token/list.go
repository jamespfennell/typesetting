package token

// List represents a token list, a fundamental data type in TeX.
// A token list is an ordered collection of Token types which are retrieved on demand.
//
// The name 'token list' originates in the TeX82 implementation.
// Conceptually it's better to think of a token list as a stream instead of a list:
// in many cases the contents of the list is not predetermined when the list is created.
type List interface {
	// NextToken retrieves the next Token in the List.
	NextToken() (Token, error)
}

func NewSingletonList(token Token) List {
	return &singletonList{token: token}
}

type singletonList struct {
	token Token
}

func (list *singletonList) NextToken() (Token, error) {
	token := list.token
	list.token = nil
	return token, nil
}

func NewSliceBasedList(tokens []Token) List {
	return &sliceBasedList{tokens: tokens}
}

type sliceBasedList struct {
	tokens []Token
}

func (list *sliceBasedList) NextToken() (Token, error) {
	if len(list.tokens) == 0 {
		return nil, nil
	}
	token := list.tokens[0]
	list.tokens = list.tokens[1:]
	return token, nil
}

func NewChainedList(lists ...List) List {
	return &chainedList{lists: lists}
}

type chainedList struct {
	lists []List
}

func (list *chainedList) NextToken() (Token, error) {
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

func NewObservedList(list List, observationChan chan<- Token) List {
	return &observedList{list: list, observationChan: observationChan}
}

type observedList struct {
	list            List
	observationChan chan<- Token
}

func (list *observedList) NextToken() (Token, error) {
	token, err := list.list.NextToken()
	if err != nil || token == nil {
		close(list.observationChan)
	} else {
		list.observationChan <- token
	}
	return token, err
}
