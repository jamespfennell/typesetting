// Package datastructures defines some data structures used in multiple places in GoTex.
package datastructures

// ScopedMap is a map data structure that has a notion of scopes. All changes made to the map during a scope
// are rolled back at the end of the scope. Changes made by a scope are inherited by any sub-scopes created within
// the current scope.
//
//	m := NewScopedMap()
//	m.Set("key", "first value")
//	m.Get("key")  // will be equal to "first value"
//	m.BeginScope()
//	m.Set("key", "second value")
//	m.Get("key")  // will be equal to "second value"
//	m.EndScope()
//	m.Get("key")  // will be equal to "first value"
//
// The implementation is such that all operations are O(1).
type ScopedMap struct {
	keyToRootNode    map[string]*scopedMapNode
	changedKeysStack []map[string]bool
}

// NewScopedMap creates a new scoped map.
func NewScopedMap() ScopedMap {
	changedKeysStack := make([]map[string]bool, 1)
	changedKeysStack[0] = make(map[string]bool)
	return ScopedMap{
		keyToRootNode:    make(map[string]*scopedMapNode),
		changedKeysStack: changedKeysStack,
	}
}

type scopedMapNode struct {
	value    interface{}
	nextNode *scopedMapNode
}

// BeginScope starts a new scope within the map.
func (scopedMap *ScopedMap) BeginScope() {
	scopedMap.changedKeysStack = append(scopedMap.changedKeysStack, make(map[string]bool))
}

// EndScope ends the current scope and rolls back all changes. Note this function will panic if no scope currently
// exists.
func (scopedMap *ScopedMap) EndScope() {
	if len(scopedMap.changedKeysStack) <= 1 {
		panic("Cannot end scope - no scope currently exists!")
	}
	for key := range scopedMap.currentScopeChangedKeys() {
		scopedMap.keyToRootNode[key] = scopedMap.keyToRootNode[key].nextNode
	}
	scopedMap.changedKeysStack = scopedMap.changedKeysStack[:len(scopedMap.changedKeysStack)-1]
}

// Set sets the value of a key.
func (scopedMap *ScopedMap) Set(key string, value interface{}) {
	if scopedMap.currentScopeChangedKeys()[key] {
		scopedMap.keyToRootNode[key].value = value
		return
	}
	node := scopedMapNode{
		value:    value,
		nextNode: scopedMap.keyToRootNode[key],
	}
	scopedMap.keyToRootNode[key] = &node
	// Don't bother adding to the changed keys set if this is the global scope, as we'll never use these keys.
	if len(scopedMap.changedKeysStack) > 1 {
		scopedMap.currentScopeChangedKeys()[key] = true
	}
}

// Get retrieves the value of a key.
// TODO: return interface{}, bool
func (scopedMap *ScopedMap) Get(key string) interface{} {
	node := scopedMap.keyToRootNode[key]
	if node == nil {
		return nil
	}
	return node.value
}

func (scopedMap *ScopedMap) currentScopeChangedKeys() map[string]bool {
	return scopedMap.changedKeysStack[len(scopedMap.changedKeysStack)-1]
}
