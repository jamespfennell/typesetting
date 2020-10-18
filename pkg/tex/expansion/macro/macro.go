package macro

import (
	"errors"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

type Macro struct {
	startingTokens []token.Token
	parameterTemplates []parameterTemplate
	replacementTemplates []replacementTemplate
}

func (m *Macro) addParameterToken(t token.Token) {
	if len(m.parameterTemplates) == 0 {
		m.startingTokens = append(m.startingTokens, t)
		return
	}
	m.parameterTemplates[len(m.parameterTemplates) - 1].delimiters = append(
		m.parameterTemplates[len(m.parameterTemplates) - 1].delimiters,
		t,
	)
}

type parameterTemplate struct {
	// If delimiters is empty, this is an undelimited parameterTemplate.
	delimiters []token.Token
}

type replacementTemplate struct {
	// parameterIndex is ignored for the first replacementTemplate.
	parameterIndex int
	tokens []token.Token
}

func Def(ctx *context.Context, s stream.TokenStream) ([]token.Token, error) {
	t, err := s.NextToken()
	if err != nil {
		return []token.Token{t}, err
	}
	if t == nil || ! t.IsCommand() {
		return nil, errors.New("expected command token, received something else")
	}
	return []token.Token{t}, nil
}

func (m *Macro) readArgumentText(ctx *context.Context, s stream.TokenStream) error {
	scopeDepth := 0
	m.replacementTemplates = append(m.replacementTemplates, replacementTemplate{parameterIndex: -1})
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
			intVal, ok := charToInt[t.Value()]
			if !ok {
				return errors.New("unexpected token after #")
			}
			if len(m.parameterTemplates) > intVal {
				return errors.New("not that many parameters")
			}
			m.replacementTemplates = append(m.replacementTemplates, replacementTemplate{parameterIndex: intVal})
		default:
			m.replacementTemplates[len(m.replacementTemplates) - 1].tokens = append(
				m.replacementTemplates[len(m.replacementTemplates) - 1].tokens,
				t,
			)
		}

	}
}

func (m *Macro) readParameterText(ctx *context.Context, s stream.TokenStream) error {
	lastParameter := -1
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
				return errors.New("TODO: this is not an error")
			}
			intVal, ok := charToInt[t.Value()]
			if !ok {
				return errors.New("unexpected token after #")
			}
			if intVal != lastParameter + 1 {
				return errors.New("unexpected number after #")
			}
			lastParameter++
			m.parameterTemplates = append(m.parameterTemplates, parameterTemplate{})
		default:
			m.addParameterToken(t)
		}
	}
}

var charToInt map[string]int

func init() {
	charToInt = make(map[string]int, 10)
	charToInt["0"] = 0
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

