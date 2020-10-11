package command

import (
	"fmt"
	"github.com/jamespfennell/typesetting/pkg/datastructures"
)

type Command interface {}


type Registry struct {
	scopedMap datastructures.ScopedMap
}

func NewRegistry() Registry {
	return Registry{scopedMap: datastructures.NewScopedMap()}
}

func (registry *Registry) GetCommand(name string) (Command, bool) {
	cmd := registry.scopedMap.Get(name)
	if cmd == nil {
		return nil, false
	}
	return cmd, true
}

func (registry *Registry) SetCommand(name string, cmd Command) {
	if cmd == nil {
		panic(fmt.Sprintf("Attempted to register nil command under name %q.", name))
	}
	registry.scopedMap.Set(name, cmd)
}
