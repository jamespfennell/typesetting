package context

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/command"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
)

// TODO this should be passed around by pointer
type Context struct {
	CatCodeMap *catcode.Map
	command.Registry
	TokenizerChannel chan<- token.Token
	ExpansionChannel chan<- token.Token
}
