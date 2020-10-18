package token

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
)

type Source interface {
	String() string
}

type Token interface {
	Value() string
	CatCode() catcode.CatCode
	IsCommand() bool
	Source() Source
}

type characterToken struct {
	value   string
	catCode catcode.CatCode
	source  Source
}

func (token characterToken) Value() string {
	return token.value
}

func (token characterToken) IsCommand() bool {
	return token.catCode == -1
}

func (token characterToken) CatCode() catcode.CatCode {
	return token.catCode
}

func (token characterToken) Source() Source {
	return token.source
}

func (token characterToken) String() string {
	if token.CatCode() == -1 {
		return fmt.Sprintf("[cmd: %s]", token.Value())
	}
	return fmt.Sprintf("[val: %s; cc: %d]", token.Value(), token.CatCode())
}

func NewCommandToken(value string, source Source) Token {
	return characterToken{value: value, catCode: -1, source: source}
}

func NewCharacterToken(value string, code catcode.CatCode, source Source) Token {
	return characterToken{value: value, catCode: code, source: source}
}
