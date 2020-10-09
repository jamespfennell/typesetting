package token

import "github.com/jamespfennell/typesetting/pkg/tex/catcode"

type Token interface {
	Value() string
	CatCode() catcode.CatCode
	IsCommand() bool
}

type characterToken struct {
	value   string
	catCode catcode.CatCode
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

func NewCommandToken(value string) Token {
	return characterToken{value: value, catCode: -1}
}

func NewCharacterToken(value string, code catcode.CatCode) Token {
	return characterToken{value: value, catCode: code}
}
