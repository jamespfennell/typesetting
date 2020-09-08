package knuthplass

import (
	"errors"
)

// Item is a primitive element in the Knuth-Plass typesetting algorithm.
// It is either a box, glue or penalty; however, the interface abstracts away these types and instead encourages
// consumers to interact with Items using polymorphism.
type Item interface {
	// Width returns the ideal width of the item.
	Width() int64

	// EndOfLineWidth returns the ideal width of the item, if that item appears at the end of a line.
	// Concretely, penalty items have width 0 unless they appear at the end of a line, whereas glue items have width
	// 0 when they appear at the end of a line.
	EndOfLineWidth() int64

	// Shrinkability returns a quantity by which the item may be shrunk proportional to.
	// The item will be shrunk at most one times this amount.
	Shrinkability() int64

	// Stretchability returns a quantity by which the item may be stretched proportional to.
	// The item may be stretched by many times this amount, though with a large breakpointPenalty.
	Stretchability() int64

	// BreakpointPenalty returns the additional penalty assessed if a line breaks at this item.
	BreakpointPenalty() int64

	// IsFlaggedBreakpoint returns whether a breakpoint at this item is isFlagged.
	// Two consecutive isFlagged breakpoints are assessed an additional penalty.
	IsFlaggedBreakpoint() bool

	// IsValidBreakpoint returns whether a line can break at this item.
	// Its single argument is the preceding Item in the line.
	IsValidBreakpoint(Item) bool

	// IsBox returns whether the item is a box.
	// In general use of this method should be avoided and code should instead rely on the other polymorphic
	// methods.
	IsBox() bool
}

// NewBox creates and returns a new box item.
func NewBox(width int64) Item {
	return &box{width: width}
}

type box struct {
	width int64
}

func (box *box) Width() int64 {
	return box.width
}

func (box *box) EndOfLineWidth() int64 {
	return box.width
}

func (*box) Shrinkability() int64 {
	return 0
}

func (*box) Stretchability() int64 {
	return 0
}

func (*box) BreakpointPenalty() int64 {
	return 0
}

func (*box) IsFlaggedBreakpoint() bool {
	return false
}

func (*box) IsValidBreakpoint(Item) bool {
	return false
}

func (*box) IsBox() bool {
	return true
}

// NewGlue creates and returns a new glue item.
func NewGlue(width int64, shrinkability int64, stretchability int64) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: stretchability}
}

type glue struct {
	width          int64
	shrinkability  int64
	stretchability int64
}

func (glue *glue) Width() int64 {
	return glue.width
}

func (glue *glue) EndOfLineWidth() int64 {
	return 0
}

func (glue *glue) Shrinkability() int64 {
	return glue.shrinkability
}

func (glue *glue) Stretchability() int64 {
	return glue.stretchability
}

func (*glue) BreakpointPenalty() int64 {
	return 0
}

func (*glue) IsFlaggedBreakpoint() bool {
	return false
}

func (*glue) IsValidBreakpoint(precedingItem Item) bool {
	if precedingItem == nil {
		return false
	}
	return precedingItem.IsBox()
}

func (*glue) IsBox() bool {
	return false
}

// PosInfBreakpointPenalty represents an infinitely positive penalty which makes it forbidden for the associated item
// to be a breakpoint. Any breakpoint penalty larger than this is also considered positive infinite.
const PosInfBreakpointPenalty int64 = 10000

// NegInfBreakpointPenalty represents an infinitely negative penalty which makes it mandatory for the associated item
// to be a breakpoint. Any breakpoint penalty more negative than this is also considered negative infinite.
const NegInfBreakpointPenalty int64 = -10000

// InfiniteStretchability is a stretchability constant such that the item can be stretched an arbitrary amount with
// no breakpoint penalty (or at least, no contribution to the adjustment ratio).
// Any stretchability larger than InfiniteStretchability is, itself, considered infinite.
const InfiniteStretchability int64 = 100000

// NewPenalty creates and returns a new penalty item.
func NewPenalty(width int64, breakpointPenalty int64, isFlagged bool) Item {
	if breakpointPenalty > PosInfBreakpointPenalty {
		breakpointPenalty = PosInfBreakpointPenalty
	} else if breakpointPenalty < NegInfBreakpointPenalty {
		breakpointPenalty = NegInfBreakpointPenalty
	}
	return &penalty{width: width, breakpointPenalty: breakpointPenalty, isFlagged: isFlagged}
}

type penalty struct {
	width             int64
	breakpointPenalty int64
	isFlagged         bool
}

func (penalty *penalty) Width() int64 {
	return 0
}

func (penalty *penalty) EndOfLineWidth() int64 {
	return penalty.width
}

func (*penalty) Shrinkability() int64 {
	return 0
}

func (*penalty) Stretchability() int64 {
	return 0
}

func (penalty *penalty) BreakpointPenalty() int64 {
	return penalty.breakpointPenalty
}

func (penalty *penalty) IsFlaggedBreakpoint() bool {
	return penalty.isFlagged
}

func (penalty *penalty) IsValidBreakpoint(Item) bool {
	return penalty.BreakpointPenalty() < PosInfBreakpointPenalty
}

func (*penalty) IsBox() bool {
	return false
}

// ItemList is a data structure representing an ordered list of items and common operations on them.
//
// It can be used to find the total width, shrinkability and stretchability of a list of items representing a line.
// These computations incorporate various pieces of logic; for example,
// - a glue item at the end of a lineIndex does not contribute to any of these quantities
// - a penalty item only contributes if it is at the end of a lineIndex.
// - all items before the first box in a lineIndex are ignored
// In addition, the data structure is implemented such that all of these computations are fast (constant in time).
type ItemList struct {
	aggregateWidth          []int64
	aggregateShrinkability  []int64
	aggregateStretchability []int64
	positionToNextBoxOffset []int
	items                   []Item
}

// Length returns the number of items in this list.
func (itemList *ItemList) Length() int {
	return len(itemList.items)
}

// Get returns an item in the list at a given index.
// If the index is negative nil is returned; if the index is too big, a panic will occur.
func (itemList *ItemList) Get(index int) Item {
	if index < 0 {
		return nil
	}
	return itemList.items[index]
}

// Slice returns a subsequence of the ItemList.
// It's analogous to sequence slicing both in terms of semantics and implementation.
// In terms of semantics, it returns the half open subsequence containing the endpoint a but not the endpoint b.
// In terms of implementation, the new ItemList internally just contains pointers to data within the old ItemList,
// and so slicing is memory efficient and fast.
func (itemList *ItemList) Slice(a int, b int) *ItemList {
	return &ItemList{
		aggregateWidth:          itemList.aggregateWidth[a : b+1],
		aggregateShrinkability:  itemList.aggregateShrinkability[a : b+1],
		aggregateStretchability: itemList.aggregateStretchability[a : b+1],
		positionToNextBoxOffset: itemList.positionToNextBoxOffset[a:b],
		items:                   itemList.items[a:b],
	}
}

// NewItemList constructs an ItemList from the provided slice of Items.
func NewItemList(items []Item) *ItemList {
	lineData := &ItemList{
		aggregateWidth:          make([]int64, len(items)+1),
		aggregateShrinkability:  make([]int64, len(items)+1),
		aggregateStretchability: make([]int64, len(items)+1),
		positionToNextBoxOffset: make([]int, len(items)),
		items:                   items,
	}
	lineData.aggregateWidth[0] = 0
	lineData.aggregateShrinkability[0] = 0
	lineData.aggregateStretchability[0] = 0
	for position, item := range items {
		lineData.aggregateWidth[position+1] =
			lineData.aggregateWidth[position] +
				item.Width()
		lineData.aggregateShrinkability[position+1] =
			lineData.aggregateShrinkability[position] +
				item.Shrinkability()
		lineData.aggregateStretchability[position+1] =
			lineData.aggregateStretchability[position] +
				item.Stretchability()
	}
	itemIndex := 0
	boxIndex := 0
	for boxIndex < len(items) {
		if !items[boxIndex].IsBox() {
			boxIndex++
			continue
		}
		lineData.positionToNextBoxOffset[itemIndex] = boxIndex - itemIndex
		if itemIndex == boxIndex {
			boxIndex++
		}
		itemIndex++
	}
	for itemIndex < len(items) {
		lineData.positionToNextBoxOffset[itemIndex] = -1
		itemIndex++
	}
	return lineData
}

// FirstBoxIndex returns the index of the first box in this ItemList, or an error if the ItemList contains no box.
// It is equivalent to iterating  over the list and checking IsBox for each item; however, the call is O(1).
func (itemList *ItemList) FirstBoxIndex() (int, error) {
	nextBoxIndex := itemList.positionToNextBoxOffset[0]
	if nextBoxIndex < 0 || nextBoxIndex >= len(itemList.items) {
		return -1, errors.New("this ItemList contains no boxes")
	}
	return nextBoxIndex, nil
}

// Width returns the width of a line consisting of elements of the ItemList.
func (itemList *ItemList) Width() int64 {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	return itemList.aggregateWidth[len(itemList.items)] -
		itemList.aggregateWidth[nextBoxIndex] +
		itemList.items[len(itemList.items)-1].EndOfLineWidth() -
		itemList.items[len(itemList.items)-1].Width()
}

// Shrinkability returns the shrinkability of a line consisting of elements of the ItemList.
func (itemList *ItemList) Shrinkability() int64 {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	return itemList.aggregateShrinkability[len(itemList.items)] -
		itemList.aggregateShrinkability[nextBoxIndex] -
		itemList.items[len(itemList.items)-1].Shrinkability()
}

// Stretchability returns the stretchability of a line consisting of elements of the ItemList.
func (itemList *ItemList) Stretchability() int64 {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	rawStretchability := itemList.aggregateStretchability[len(itemList.items)] -
		itemList.aggregateStretchability[nextBoxIndex] -
		itemList.items[len(itemList.items)-1].Stretchability()
	if rawStretchability < InfiniteStretchability {
		return rawStretchability
	}
	return InfiniteStretchability
}
