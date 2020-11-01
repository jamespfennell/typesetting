package macro

import (
	"errors"
	"github.com/jamespfennell/typesetting/pkg/tex/context"
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

func (b *command) Invoke(ctx *context.Context, es stream.ExpandingStream) error {
	s := es.SourceStream()
	t, err := s.NextToken()
	if err != nil {
		return err
	}
	if t == nil || !t.IsCommand() {
		return errors.New("expected command token, received something else")
	}
	m := &macro{}
	m.argumentsTemplate, err = buildArgumentsTemplate(s)
	if err != nil {
		return err
	}
	m.replacement, err = buildReplacementTokens(s)
	if err != nil {
		return err
	}
	expansion.Register(ctx, t.Value(), m)
	return nil
}

func buildArgumentsTemplate(s stream.TokenStream) (argumentsTemplate, error) {
	template := argumentsTemplate{}
	lastParameter := 0
	err := func() error {
		for {
			t, err := s.NextToken()
			if err != nil {
				return err
			}
			if t == nil {
				return errors.New("unexpected end of tokenization while parsing macro argumentsTemplate")
			}
			switch t.CatCode() {
			case catcode.BeginGroup:
				return nil
			case catcode.EndGroup:
				return errors.New("unexpected end of group character in macro argumentsTemplate definition")
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
					addTokenToArgumentTemplate(&template, t)
					template.finalToken = t
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
				template.delimiters = append(template.delimiters, []token.Token{})
			default:
				addTokenToArgumentTemplate(&template, t)
			}
		}
	}()
	return template, err
}

func addTokenToArgumentTemplate(a *argumentsTemplate, t token.Token) {
	if len(a.delimiters) == 0 {
		a.prefix = append(a.prefix, t)
		return
	}
	a.delimiters[len(a.delimiters)-1] = append(
		a.delimiters[len(a.delimiters)-1],
		t,
	)
}

func buildReplacementTokens(s stream.TokenStream) (*replacementTokens, error) {
	scopeDepth := 0
	root := &replacementTokens{}
	curTokens := root
	for {
		t, err := s.NextToken()
		if err != nil {
			return nil, err
		}
		if t == nil {
			return nil, errors.New("unexpected end of tokenization while parsing macro argumentsTemplate")
		}
		switch t.CatCode() {
		case catcode.BeginGroup:
			scopeDepth += 1
		case catcode.EndGroup:
			if scopeDepth == 0 {
				return root, nil
			}
			scopeDepth -= 1
		case catcode.Parameter:
			t, err := s.NextToken()
			if err != nil {
				return nil, err
			}
			if t == nil {
				return nil, errors.New("expected command token, received something else")
			}
			if t.CatCode() == catcode.Parameter {
				curTokens.tokens = append(curTokens.tokens, t)
				continue
			}
			index, ok := charToInt[t.Value()]
			if !ok {
				return nil, errors.New("unexpected token after #: " + t.Value())
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
