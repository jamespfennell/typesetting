package catcode

import (
	"github.com/jamespfennell/typesetting/pkg/datastructures"
)

type CatCode int

const (
	Escape CatCode = iota
	BeginGroup
	EndGroup
	MathShift
	AlignmentTab
	EndOfLine
	Parameter
	Superscript
	Subscript
	Ignored
	Space
	Letter
	Other
	Active
	Comment
	Invalid
)

func (catCode CatCode) String() string {
	switch catCode {
	case Escape:
		return "escape"
	case BeginGroup:
		return "begin group"
	case Letter:
		return "letter"
	}
	return "unknown"
}

// Map is a typed version of datastructures.ScopedMap in which the values are of type CatCode.
type Map struct {
	scopedMap datastructures.ScopedMap
}

func NewCatCodeMap() Map {
	return Map{
		scopedMap: datastructures.NewScopedMap(),
	}
}

func (catCodeMap *Map) BeginScope() {
	catCodeMap.scopedMap.BeginScope()
}

func (catCodeMap *Map) EndScope() {
	catCodeMap.scopedMap.EndScope()
}

func (catCodeMap *Map) Set(key string, value CatCode) {
	catCodeMap.scopedMap.Set(key, value)
}

func (catCodeMap *Map) Get(key string) CatCode {
	value := catCodeMap.scopedMap.Get(key)
	if value == nil {
		return Other
	}
	return value.(CatCode)
}

var texDefaults map[string]CatCode = map[string]CatCode{
	"\\": Escape,
	"{":  BeginGroup,
	"}":  EndGroup,
	"$":  MathShift,
	"&":  AlignmentTab,
	"\n": EndOfLine,
	"#":  Parameter,
	"^":  Superscript,
	"_":  Subscript,
	"~":  Active,
	"%":  Comment,

	" ": Space, // TODO: other white space characters?

	"A": Letter,
	"B": Letter,
	"C": Letter,
	"D": Letter,
	"E": Letter,
	"F": Letter,
	"G": Letter,
	"H": Letter,
	"I": Letter,
	"J": Letter,
	"K": Letter,
	"L": Letter,
	"M": Letter,
	"N": Letter,
	"O": Letter,
	"P": Letter,
	"Q": Letter,
	"R": Letter,
	"S": Letter,
	"T": Letter,
	"U": Letter,
	"V": Letter,
	"W": Letter,
	"X": Letter,
	"Y": Letter,
	"Z": Letter,

	"a": Letter,
	"b": Letter,
	"c": Letter,
	"d": Letter,
	"e": Letter,
	"f": Letter,
	"g": Letter,
	"h": Letter,
	"i": Letter,
	"j": Letter,
	"k": Letter,
	"l": Letter,
	"m": Letter,
	"n": Letter,
	"o": Letter,
	"p": Letter,
	"q": Letter,
	"r": Letter,
	"s": Letter,
	"t": Letter,
	"u": Letter,
	"v": Letter,
	"w": Letter,
	"x": Letter,
	"y": Letter,
	"z": Letter,
}

func NewCatCodeMapWithTexDefaults() Map {
	m := NewCatCodeMap()
	for key, val := range texDefaults {
		m.Set(key, val)
	}
	return m
}
