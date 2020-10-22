package macro

import (
	"errors"
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

type Macro struct {
	argument
	replacement *replacementTokens
	lastToken token.Token
}

type argument struct {
	prefix     []token.Token
	delimiters [][]token.Token
}

func (a *argument) addArgumentToken(t token.Token) {
	if len(a.delimiters) == 0 {
		a.prefix = append(a.prefix, t)
		return
	}
	a.delimiters[len(a.delimiters)-1] = append(
		a.delimiters[len(a.delimiters)-1],
		t,
	)
}

type replacementTokens struct {
	tokens []token.Token
	next   *replacementParameter
}

type replacementParameter struct {
	index int // TODO: rename to make clean this is 1-based, or change to 0-based
	next  *replacementTokens
}

func Def(ctx *context.Context, s stream.TokenStream) ([]token.Token, error) {
	t, err := s.NextToken()
	if err != nil {
		return []token.Token{t}, err
	}
	if t == nil || !t.IsCommand() {
		return nil, errors.New("expected command token, received something else")
	}
	m := &Macro{}
	m.replacement = &replacementTokens{}
	if err := m.parseArgumentTokens(ctx, s); err != nil {
		return nil, err
	}
	if err := m.parseReplacementTokens(ctx, s); err != nil {
		return nil, err
	}
	expansion.Register(&ctx.Registry, t.Value(), func(ctx *context.Context, s stream.TokenStream) stream.TokenStream {
		p, err := m.matchParameters(s)
		if err != nil {
			return stream.NewErrorStream(err)
		}
		return m.output(p)
	})
	return []token.Token{}, nil
}

func (m *Macro) parseReplacementTokens(ctx *context.Context, s stream.TokenStream) error {
	scopeDepth := 0
	curTokens := m.replacement
	for {
		t, err := s.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("unexpected end of input while parsing macro argument")
		}
		switch t.CatCode() {
		case catcode.BeginGroup:
			scopeDepth += 1
		case catcode.EndGroup:
			if scopeDepth == 0 {
				return nil
			}
			scopeDepth -= 1
		case catcode.Parameter:
			t, err := s.NextToken()
			if err != nil {
				return err
			}
			if t == nil {
				return errors.New("expected command token, received something else")
			}
			if t.CatCode() == catcode.Parameter {
				curTokens.tokens = append(curTokens.tokens, t)
				continue
			}
			index, ok := charToInt[t.Value()]
			if !ok {
				return errors.New("unexpected token after #: " + t.Value())
			}
			parameter := replacementParameter{
				index: index,
				next:  &replacementTokens{},
			}
			curTokens.next = &parameter
			curTokens = parameter.next
			continue
		}
		curTokens.tokens = append(curTokens.tokens, t)
	}
}

func (m *Macro) parseArgumentTokens(ctx *context.Context, s stream.TokenStream) error {
	lastParameter := 0
	for {
		t, err := s.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("unexpected end of input while parsing macro argument")
		}
		switch t.CatCode() {
		case catcode.BeginGroup:
			return nil
		case catcode.EndGroup:
			return errors.New("unexpected end of group character in macro argument definition")
		case catcode.Parameter:
			t, err := s.NextToken()
			if err != nil {
				return err
			}
			if t == nil {
				return errors.New("expected command token, received something else")
			}
			if t.CatCode() == catcode.BeginGroup {
				// end the group according to the special #{ rule
				// return errors.New("TODO: this is not an error")
				m.addArgumentToken(t)
				m.lastToken = t
				return nil
			}
			intVal, ok := charToInt[t.Value()]
			if !ok {
				return errors.New("unexpected token after #: " + t.Value())
			}
			if intVal != lastParameter+1 {
				return errors.New("unexpected number after #: " + t.Value())
			}
			lastParameter++
			m.argument.delimiters = append(m.argument.delimiters, []token.Token{})
		default:
			m.addArgumentToken(t)
		}
	}
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

func (m *Macro) output(p *matchedParameters) stream.TokenStream {
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
	if m.lastToken != nil {
		components = append(components, stream.NewSliceStream([]token.Token{m.lastToken}))
	}
	return stream.NewChainedStream(components...)
}

func (m *Macro) matchParameters(s stream.TokenStream) (*matchedParameters, error) {
	if err := m.matchArgumentPrefix(s); err != nil {
		return nil, err
	}
	p := &matchedParameters{}
	index := 0
	for {
		if len(m.argument.delimiters) <= index {
			return p, nil
		}
		var thisParameter []token.Token
		var err error
		delimiter := m.argument.delimiters[index]
		if len(delimiter) == 0 {
			thisParameter, err = getUndelimitedParameter(s)
		} else {
			thisParameter, err = getDelimitedParameter(s, delimiter)
		}
		if err != nil {
			return nil, err
		}
		p.parameters = append(p.parameters, thisParameter)
		index++
	}
}

func getDelimitedParameter(s stream.TokenStream, delimiter []token.Token) ([]token.Token, error) {
	var seen []token.Token
	depth := 0
	for {
		if depth == 0 && tokenListHasTail(seen, delimiter) {
			return seen[:len(seen)-len(delimiter)], nil
		}
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New("unexpected end of input")
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
		return nil, errors.New("unexpected end of input")
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
			return nil, errors.New("unexpected end of input")
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

func (m *Macro) matchArgumentPrefix(s stream.TokenStream) error {
	i := 0
	for {
		if len(m.argument.prefix) <= i {
			return nil
		}
		tokenToMatch := m.argument.prefix[i]
		t, err := s.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return errors.New("unexpected end of input")
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
