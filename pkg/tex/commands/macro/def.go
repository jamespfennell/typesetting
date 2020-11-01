package macro

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
	"github.com/jamespfennell/typesetting/pkg/tex/errors"
	"github.com/jamespfennell/typesetting/pkg/tex/expansion"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

// command is used to build!
type command struct {
	global    bool
	outer     bool
	preExpand bool
	ready     bool
}

func GetDef() context.ExecutionCommand {
	return &command{ready: true}
}

const determiningMacroDefinitionTarget = "determining the control sequence being defined in a macro definition"

func (b *command) Invoke(ctx *context.Context, es stream.ExpandingStream) error {
	s := es.SourceStream()
	t, err := s.NextToken()
	if err != nil {
		return err
	}
	if t == nil {
		return errors.NewUnexpectedEndOfInputError(determiningMacroDefinitionTarget)
	}
	if !t.IsCommand() {
		return errors.NewUnexpectedTokenError(
			t,
			"command token",
			"a non-command token with value "+t.Value(),
			determiningMacroDefinitionTarget)
	}
	m := &macro{}
	var replacementEndToken token.Token
	m.argument, replacementEndToken, err = buildArgumentsTemplate(s)
	if err != nil {
		return err
	}
	m.replacement, err = buildReplacementTokens(s, replacementEndToken, len(m.argument.delimiters))
	if err != nil {
		return err
	}
	expansion.Register(ctx, t.Value(), m)
	return nil
}

const parsingArgumentTemplate = "parsing argument template in macro definition"

func buildArgumentsTemplate(s stream.TokenStream) (argumentTemplate, token.Token, error) {
	template := argumentTemplate{}
	var replacementEndToken token.Token
	curParameterIndex := 0
	err := func() error {
		for {
			t, err := s.NextToken()
			if err != nil {
				return err
			}
			if t == nil {
				return errors.NewUnexpectedEndOfInputError(parsingArgumentTemplate)
			}
			switch t.CatCode() {
			case catcode.BeginGroup:
				return nil
			case catcode.EndGroup:
				return errors.NewUnexpectedTokenError(t, "", t.Description(), parsingArgumentTemplate)
			case catcode.Parameter:
				t, err := s.NextToken()
				if err != nil {
					return err
				}
				if t == nil {
					return errors.NewUnexpectedEndOfInputError(parsingArgumentTemplate)
				}
				if t.CatCode() == catcode.BeginGroup {
					// end the group according to the special #{ rule
					addTokenToArgumentTemplate(&template, t)
					replacementEndToken = t
					return nil
				}
				parameterIndex, ok := charToParameterIndex[t.Value()]
				if !ok {
					return errors.NewUnexpectedTokenError(
						t,
						"a number between 1 and 9 indicating which parameter is being referred to",
						"the value "+t.Value(),
						parsingArgumentTemplate)
				}
				if parameterIndex != curParameterIndex {
					return errors.NewUnexpectedTokenError(
						t,
						fmt.Sprintf(
							"the number %[1]d because this is parameter number %[1]d to appear",
							curParameterIndex+1),
						"the number "+t.Value(),
						parsingArgumentTemplate)
				}
				curParameterIndex++
				template.delimiters = append(template.delimiters, []token.Token{})
			default:
				addTokenToArgumentTemplate(&template, t)
			}
		}
	}()
	return template, replacementEndToken, err
}

func addTokenToArgumentTemplate(a *argumentTemplate, t token.Token) {
	if len(a.delimiters) == 0 {
		a.prefix = append(a.prefix, t)
		return
	}
	a.delimiters[len(a.delimiters)-1] = append(
		a.delimiters[len(a.delimiters)-1],
		t,
	)
}

func buildReplacementTokens(s stream.TokenStream, finalToken token.Token, numParams int) (*replacementTokens, error) {
	scopeDepth := 0
	root := &replacementTokens{}
	curTokens := root
	for {
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.NewUnexpectedEndOfInputError(parsingArgumentTemplate)
		}
		switch t.CatCode() {
		case catcode.BeginGroup:
			scopeDepth += 1
		case catcode.EndGroup:
			if scopeDepth == 0 {
				if finalToken != nil {
					curTokens.tokens = append(curTokens.tokens, finalToken)
				}
				return root, nil
			}
			scopeDepth -= 1
		case catcode.Parameter:
			t, err := s.NextToken()
			if err != nil {
				return nil, err
			}
			if t == nil {
				return nil, errors.NewUnexpectedEndOfInputError(parsingArgumentTemplate)
			}
			if t.CatCode() == catcode.Parameter {
				curTokens.tokens = append(curTokens.tokens, t)
				continue
			}
			index, ok := charToParameterIndex[t.Value()]
			if !ok {
				return nil, errors.NewUnexpectedTokenError(
					t,
					fmt.Sprintf(
						"a number between 1 and %d indicating which parameter is being referred to",
						numParams),
					"the value "+t.Value(),
					parsingArgumentTemplate)
			}
			if index >= numParams {
				var msg string
				switch numParams {
				case 0:
					msg = "no parameter token because this macro has 0 parameters"
				case 1:
					msg = "the number 1 because this macro has only 1 parameter"
				default:
					msg = fmt.Sprintf(
						"a number between 1 and %[1]d inclusive because this macro has only %[1]d parameters",
						numParams)
				}
				return nil, errors.NewUnexpectedTokenError(t, msg, "the number "+t.Value(), parsingArgumentTemplate)
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
