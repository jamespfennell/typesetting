package macro

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

type macro struct {
	argumentsTemplate
	replacement *replacementTokens
}

func (m macro) Invoke(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	p, err := m.matchParameters(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	for i, param := range p.parameters {
		// TODO: make this loggable through the logger in the context
		// fmt.Println(fmt.Sprintf("#%d<-%v", i, param))
		_ = i
		_ = param
	}
	return m.output(p)
}

type argumentsTemplate struct {
	prefix     []token.Token
	delimiters [][]token.Token
	finalToken token.Token
}

type replacementTokens struct {
	tokens []token.Token
	next   *replacementParameter
}

type replacementParameter struct {
	index int // TODO: rename to make clean this is 1-based, or change to 0-based
	next  *replacementTokens
}

var charToInt map[string]int

func init() {
	charToInt = make(map[string]int, 10)
	charToInt["1"] = 1
	charToInt["2"] = 2
	charToInt["3"] = 3
	charToInt["4"] = 4
	charToInt["5"] = 5
	charToInt["6"] = 6
	charToInt["7"] = 7
	charToInt["8"] = 8
	charToInt["9"] = 9
}

type matchedParameters struct {
	parameters [][]token.Token // todo: rename values
}

func (m *macro) output(p *matchedParameters) stream.TokenStream {
	// TODO: may me more efficient if components is just a slice
	//  we can pre allocate it b/c we can calculate the size of the output :)
	var components []stream.TokenStream
	replacementTokens := m.replacement
	for {
		if replacementTokens == nil {
			fmt.Println("Ending")
			break
		}
		components = append(components, stream.NewSliceStream(replacementTokens.tokens))
		if replacementTokens.next == nil {
			break
		}
		parameter := replacementTokens.next.index
		components = append(components, stream.NewSliceStream(p.parameters[parameter-1]))
		replacementTokens = replacementTokens.next.next
	}
	if m.argumentsTemplate.finalToken != nil {
		components = append(components, stream.NewSliceStream([]token.Token{m.argumentsTemplate.finalToken}))
	}
	return stream.NewChainedStream(components...)
}

func (m *macro) matchParameters(s stream.TokenStream) (*matchedParameters, error) {
	if err := m.matchArgumentPrefix(s); err != nil {
		return nil, err
	}
	p := &matchedParameters{}
	index := 0
	for {
		if len(m.argumentsTemplate.delimiters) <= index {
			return p, nil
		}
		var thisParameter []token.Token
		var err error
		delimiter := m.argumentsTemplate.delimiters[index]
		if len(delimiter) == 0 {
			thisParameter, err = getUndelimitedParameter(s)
		} else {
			thisParameter, err = getDelimitedParameter(s, delimiter, index+1)
		}
		if err != nil {
			return nil, err
		}
		p.parameters = append(p.parameters, thisParameter)
		index++
	}
}

func getDelimitedParameter(s stream.TokenStream, delimiter []token.Token, paramNum int) ([]token.Token, error) {
	var seen []token.Token
	depth := 0
	targetDepth := 0
	// Hack to handle the special case of ending begin group
	if delimiter[len(delimiter)-1].CatCode() == catcode.BeginGroup {
		targetDepth = 1
	}
	for {
		if depth == targetDepth && tokenListHasTail(seen, delimiter) {
			preResult := seen[:len(seen)-len(delimiter)]
			if len(preResult) > 1 {
				if (preResult[0].CatCode() == catcode.BeginGroup) && (preResult[len(preResult)-1].CatCode() == catcode.EndGroup) {
					return preResult[1 : len(preResult)-1], nil
				}
			}
			return seen[:len(seen)-len(delimiter)], nil
		}
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			fmt.Println(seen)
			return nil, errors.New(
				fmt.Sprintf("unexpected end of tokenization while reading parameter %d of macro \\", paramNum),
			)
			// TODO: tokenization ended here
			// TODO: macro invocation starts here
			// TODO: macro defined here
		}
		if t.CatCode() == catcode.BeginGroup {
			depth += 1
		}
		if t.CatCode() == catcode.EndGroup {
			depth -= 1
		}
		seen = append(seen, t)
	}
}

func tokenListHasTail(list, tail []token.Token) bool {
	headLen := len(list) - len(tail)
	if headLen < 0 {
		return false
	}
	for i, _ := range tail {
		if list[headLen+i].Value() != tail[i].Value() {
			return false
		}
		if list[headLen+i].CatCode() != tail[i].CatCode() {
			return false
		}
	}
	return true
}

func getUndelimitedParameter(s stream.TokenStream) ([]token.Token, error) {
	t, err := s.NextToken()
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.New("unexpected end of tokenization")
	}
	if t.CatCode() != catcode.BeginGroup {
		return []token.Token{t}, nil
	}
	var result []token.Token
	scope := 0
	for {
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New("unexpected end of tokenization")
		}
		if t.CatCode() == catcode.BeginGroup {
			scope += 1
		}
		if t.CatCode() == catcode.EndGroup {
			if scope == 0 {
				return result, nil
			}
			scope -= 1
		}
		result = append(result, t)
	}
}

func (m *macro) matchArgumentPrefix(s stream.TokenStream) error {
	i := 0
	for {
		if len(m.argumentsTemplate.prefix) <= i {
			return nil
		}
		tokenToMatch := m.argumentsTemplate.prefix[i]
		t, err := s.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("unexpected end of tokenization")
		}
		if tokenToMatch.Value() != t.Value() {
			return errors.New(fmt.Sprintf("unexpected token value %s; expected %s", t.Value(), tokenToMatch.Value()))
		}
		if tokenToMatch.CatCode() != t.CatCode() {
			return errors.New("unexpected token cat code")
		}
		i++
	}
}
