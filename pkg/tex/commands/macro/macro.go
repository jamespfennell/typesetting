package macro

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/errors"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

type macro struct {
	argument    argumentTemplate
	replacement *replacementTokens
}

func (m macro) Invoke(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
	p, err := m.argument.buildParameterValues(s)
	if err != nil {
		return stream.NewErrorStream(err)
	}
	for i, param := range p {
		// TODO: make this loggable through the logger in the context
		// fmt.Println(fmt.Sprintf("#%d<-%v", i, param))
		_ = i
		_ = param
	}
	return m.replacement.performReplacement(p)
}

type argumentTemplate struct {
	prefix     []token.Token
	delimiters [][]token.Token
}

type replacementTokens struct {
	tokens []token.Token
	next   *replacementParameter
}

type replacementParameter struct {
	index int
	next  *replacementTokens
}

type parameterValue []token.Token

var charToParameterIndex map[string]int

func init() {
	charToParameterIndex = make(map[string]int, 9)
	charToParameterIndex["1"] = 0
	charToParameterIndex["2"] = 1
	charToParameterIndex["3"] = 2
	charToParameterIndex["4"] = 3
	charToParameterIndex["5"] = 4
	charToParameterIndex["6"] = 5
	charToParameterIndex["7"] = 6
	charToParameterIndex["8"] = 7
	charToParameterIndex["9"] = 8
}

func (replacement *replacementTokens) performReplacement(parameterValues []parameterValue) stream.TokenStream {
	outputSize := 0
	for _, v := range parameterValues {
		outputSize += len(v)
	}
	r := replacement
	for r != nil {
		outputSize += len(r.tokens)
		if r.next == nil {
			break
		}
		r = r.next.next
	}
	output := make([]token.Token, 0, outputSize)
	for {
		if replacement == nil {
			break
		}
		output = append(output, replacement.tokens...)
		if replacement.next == nil {
			break
		}
		parameterIndex := replacement.next.index
		output = append(output, parameterValues[parameterIndex]...)
		//goland:noinspection GoAssignmentToReceiver
		replacement = replacement.next.next
	}
	return stream.NewSliceStream(output)
}

func (a *argumentTemplate) buildParameterValues(s stream.TokenStream) ([]parameterValue, error) {
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
			value, err = buildUndelimitedParameterValue(s)
		} else {
			value, err = buildDelimitedParameterValue(s, delimiter, index+1)
		}
		if err != nil {
			return nil, err
		}
		p = append(p, value)
		index++
	}
}

func buildDelimitedParameterValue(s stream.TokenStream, delimiter []token.Token, paramNum int) (parameterValue, error) {
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
			return nil, errors.NewUnexpectedEndOfInputError(
				fmt.Sprintf("reading parameter number %d of macro", paramNum),
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

func buildUndelimitedParameterValue(s stream.TokenStream) (parameterValue, error) {
	t, err := s.NextToken()
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errors.NewUnexpectedEndOfInputError("reading parameter value")
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
			return nil, errors.NewUnexpectedEndOfInputError("reading parameter value")
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

const readingArgumentPrefix = "matching the prefix of a macro argument"

func (a *argumentTemplate) consumePrefix(s stream.TokenStream) error {
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
			return errors.NewUnexpectedEndOfInputError(readingArgumentPrefix)
		}
		if tokenToMatch.Value() != t.Value() || tokenToMatch.CatCode() != t.CatCode() {
			return errors.NewUnexpectedTokenError(
				t,
				tokenToMatch.Description(),
				t.Description(),
				readingArgumentPrefix)
		}
		i++
	}
}
