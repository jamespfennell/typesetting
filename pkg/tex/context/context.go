package context

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/datastructures"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

type Context struct {
	Expansion struct {
		Commands ExpansionCommandMap
		Log      logging.LogSender
	}
	Tokenization struct {
		CatCodes catcode.Map
		Log      logging.LogSender
	}
}

func NewContext() *Context {
	ctx := Context{}
	ctx.Expansion.Commands = NewExpansionCommandMap()

	ctx.Tokenization.CatCodes = catcode.NewCatCodeMap()
	return &ctx
}

func (ctx *Context) BeginScope() {
	ctx.Expansion.Commands.m.BeginScope()
}

func (ctx *Context) EndScope() {
	ctx.Expansion.Commands.m.EndScope()
}

type ExpansionCommand interface {
	Invoke(ctx *Context, s stream.TokenStream) stream.TokenStream
}

type ExpansionCommandMap struct {
	m datastructures.ScopedMap
}

func NewExpansionCommandMap() ExpansionCommandMap {
	return ExpansionCommandMap{m: datastructures.NewScopedMap()}
}

func (m *ExpansionCommandMap) Get(name string) (ExpansionCommand, bool) {
	cmd := m.m.Get(name)
	if cmd == nil {
		return nil, false
	}
	return cmd.(ExpansionCommand), true
}

func (m *ExpansionCommandMap) Set(name string, cmd ExpansionCommand) {
	if cmd == nil {
		panic(fmt.Sprintf("Attempted to register nil command under name %q.", name))
	}
	m.m.Set(name, cmd)
}
