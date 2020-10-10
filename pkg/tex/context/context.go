package context

import (
	"github.com/jamespfennell/typesetting/pkg/datastructures"
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
)

type Context struct {
	CatCodeMap       *catcode.Map
	CommandMap       *CommandMap
	TokenizerChannel chan<- token.Token
	ExpansionChannel chan<- token.Token
}
type Command interface {
	// Execute(ctx tex.Context, list token.TokenStream) (token.TokenStream, error)
	// IsExpansionCommand() bool
}

type ExpansionCommand func(ctx Context, list stream.TokenStream) (stream.TokenStream, error)

// TODO: should probably extract this to its own data structure and include using composition
func (ctx *Context) RegisterExpansionCommand(name string, expansionCommand ExpansionCommand) {
	ctx.CommandMap.Set(name, expansionCommand)
}

// TODO: move these to its own package and type alias to avoid duplicate code

type CommandMap struct {
	scopedMap datastructures.ScopedMap
}

func NewCommandMap() CommandMap {
	return CommandMap{
		scopedMap: datastructures.NewScopedMap(),
	}
}

func (commandMap *CommandMap) BeginScope() {
	commandMap.scopedMap.BeginScope()
}

func (commandMap *CommandMap) EndScope() {
	commandMap.scopedMap.EndScope()
}

func (commandMap *CommandMap) Set(key string, value Command) {
	commandMap.scopedMap.Set(key, value)
}

func (commandMap *CommandMap) Get(key string) Command {
	value := commandMap.scopedMap.Get(key)
	if value == nil {
		return nil
	}
	return value.(Command)
}
