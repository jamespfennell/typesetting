package datastructures

import "testing"

// Test can't cancel the Global context

func TestScopedDict_BasicCase(t *testing.T) {
	m := NewScopedMap()
	m.Set("A", "B")
	if m.Get("A") != "B" {
		t.Errorf("Recieved: %v; expected: %v", m.Get("A"), "B")
	}
	if m.Get("C") != nil {
		t.Errorf("Recieved: %v; expected: %v", m.Get("C"), nil)
	}
}

func TestScopedDict_SingleScope(t *testing.T) {
	m := NewScopedMap()
	m.Set("A", "B")
	m.NewScope()
	m.Set("A", "D")
	m.Set("A", "C")
	if m.Get("A") != "C" {
		t.Errorf("Recieved: %v; expected: %v", m.Get("A"), "C")
	}
	m.EndScope()
	if m.Get("A") != "B" {
		t.Errorf("Recieved: %v; expected: %v", m.Get("A"), "B")
	}
}


func TestScopedDict_TwoSingleScopes(t *testing.T) {
	m := NewScopedMap()
	m.Set("A", "B")
	m.NewScope()
	m.Set("A", "C")
	if m.Get("A") != "C" {
		t.Errorf("Recieved: %v; expected: %v", m.Get("A"), "C")
	}
	m.EndScope()
	m.NewScope()
	if m.Get("A") != "B" {
		t.Errorf("Recieved: %v; expected: %v", m.Get("A"), "B")
	}
}

