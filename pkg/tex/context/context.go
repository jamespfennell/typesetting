package context

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/command"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
)

type Context struct {
	CatCodeMap catcode.Map
	command.Registry
	TokenizerLog logging.LogSender
	ExpansionLog logging.LogSender
}

func NewContext() *Context {
	return &Context{
		CatCodeMap: catcode.NewCatCodeMap(),
		Registry:   command.NewRegistry(),
	}
}
