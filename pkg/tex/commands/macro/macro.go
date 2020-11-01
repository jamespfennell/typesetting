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
	p, err := m.calculateParameterValues(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	for i, param := range p {
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

type parameterValue []token.Token

var charToInt map[string]int

func init() {
	charToInt = make(map[string]int, 9)
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

func (m *macro) output(p []parameterValue) stream.TokenStream {
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
		components = append(components, stream.NewSliceStream(p[parameter-1]))
		replacementTokens = replacementTokens.next.next
	}
	if m.argumentsTemplate.finalToken != nil {
		components = append(components, stream.NewSliceStream([]token.Token{m.argumentsTemplate.finalToken}))
	}
	return stream.NewChainedStream(components...)
}

func (a *argumentsTemplate) calculateParameterValues(s stream.TokenStream) ([]parameterValue, error) {
	if err := a.consumePrefix(s); err != nil {
		return nil, err
	}
	var p []parameterValue
	index := 0
	for {
		if len(a.delimiters) <= index {
			return p, nil
		}
		var value parameterValue
		var err error
		delimiter := a.delimiters[index]
		if len(delimiter) == 0 {
			value, err = consumeUndelimitedParameterValue(s)
		} else {
			value, err = consumeDelimitedParameterValue(s, delimiter, index+1)
		}
		if err != nil {
			return nil, err
		}
		p = append(p, value)
		index++
	}
}

func consumeDelimitedParameterValue(s stream.TokenStream, delimiter []token.Token, paramNum int) (parameterValue, error) {
	var tokenList []token.Token
	scopeDepth := 0
	closingScopeDepth := 0
	// This handles the case of a macro whose argument ends with the special #{ tokens. In this special case the parsing
	// will end with a scope depth of 1, because the last token parsed will be the { and all braces before that will
	// be balanced.
	if delimiter[len(delimiter)-1].CatCode() == catcode.BeginGroup {
		closingScopeDepth = 1
	}
	for {
		if scopeDepth == closingScopeDepth && tokenListHasTail(tokenList, delimiter) {
			return trimOuterBracesIfPresent(tokenList[:len(tokenList)-len(delimiter)]), nil
		}
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New(
				fmt.Sprintf("unexpected end of tokenization while reading parameter %d of macro \\", paramNum),
			)
		}
		if t.CatCode() == catcode.BeginGroup {
			scopeDepth += 1
		}
		if t.CatCode() == catcode.EndGroup {
			scopeDepth -= 1
		}
		tokenList = append(tokenList, t)
	}
}

func trimOuterBracesIfPresent(list []token.Token) []token.Token {
	if len(list) <= 0 {
		return list
	}
	if (list[0].CatCode() != catcode.BeginGroup) || (list[len(list)-1].CatCode() != catcode.EndGroup) {
		return list
	}
	return list[1 : len(list)-1]
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

func consumeUndelimitedParameterValue(s stream.TokenStream) (parameterValue, error) {
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
	scopeDepth := 0
	for {
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New("unexpected end of tokenization")
		}
		if t.CatCode() == catcode.BeginGroup {
			scopeDepth += 1
		}
		if t.CatCode() == catcode.EndGroup {
			if scopeDepth == 0 {
				return result, nil
			}
			scopeDepth -= 1
		}
		result = append(result, t)
	}
}

func (a *argumentsTemplate) consumePrefix(s stream.TokenStream) error {
	i := 0
	for {
		if len(a.prefix) <= i {
			return nil
		}
		tokenToMatch := a.prefix[i]
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
