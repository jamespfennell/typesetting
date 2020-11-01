package context

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/datastructures"
	"github.com/jamespfennell/typesetting/pkg/tex/logging"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
)

// TODO: this should be in the root tex package
type Context struct {
	Expansion struct {
		Commands ExpansionCommandMap
		Log      logging.LogSender
	}
	Tokenization struct {
		CatCodes catcode.Map
		Log      logging.LogSender
	}
	Execution struct {
		Commands ExecutionCommandMap
	}
}

func NewContext() *Context {
	ctx := Context{}
	ctx.Expansion.Commands = NewExpansionCommandMap()
	ctx.Execution.Commands = NewExecutionCommandMap()

	ctx.Tokenization.CatCodes = catcode.NewCatCodeMap()
	return &ctx
}

func (ctx *Context) allScopedDataStructures() []*datastructures.ScopedMap {
	return []*datastructures.ScopedMap{
		&ctx.Expansion.Commands.m,
		&ctx.Execution.Commands.m,
	}
}

func (ctx *Context) BeginScope() {
	for _, sm := range ctx.allScopedDataStructures() {
		sm.BeginScope()
	}
}

func (ctx *Context) EndScope() {
	for _, sm := range ctx.allScopedDataStructures() {
		sm.EndScope()
	}
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

type ExecutionCommand interface {
	Invoke(ctx *Context, s stream.ExpandingStream) error
}

type ExecutionCommandMap struct {
	m datastructures.ScopedMap
}

func NewExecutionCommandMap() ExecutionCommandMap {
	return ExecutionCommandMap{m: datastructures.NewScopedMap()}
}

func (m *ExecutionCommandMap) Get(name string) (ExecutionCommand, bool) {
	cmd := m.m.Get(name)
	if cmd == nil {
		return nil, false
	}
	return cmd.(ExecutionCommand), true
}

func (m *ExecutionCommandMap) Set(name string, cmd ExecutionCommand) {
	if cmd == nil {
		panic(fmt.Sprintf("Attempted to register nil command under name %q.", name))
	}
	m.m.Set(name, cmd)
}
