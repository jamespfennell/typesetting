package primitives

import (
	"errors"
	d "github.com/jamespfennell/typesetting/pkg/distance"
)

// Item is a primitive element in the Knuth-Plass typesetting algorithm.
// It is either a box, glue or penalty; however, the interface abstracts away these types and instead encourages
// consumers to interact with Items using polymorphism.
type Item interface {
	// Width returns the ideal width of the item.
	Width() d.Distance

	// EndOfLineWidth returns the ideal width of the item, if that item appears at the end of a line.
	// Concretely, penalty items have width 0 unless they appear at the end of a line, whereas glue items have width
	// 0 when they appear at the end of a line.
	EndOfLineWidth() d.Distance

	// Shrinkability returns a quantity by which the item may be shrunk proportional to.
	// The item will be shrunk at most one times this amount.
	Shrinkability() d.Distance

	// Stretchability returns a quantity by which the item may be stretched proportional to.
	// The item may be stretched by many times this amount, though with a large breakpointPenalty.
	Stretchability() d.Distance

	// BreakpointPenalty returns the additional penalty assessed if a line breaks at this item.
	BreakpointPenalty() PenaltyCost

	// IsFlaggedBreakpoint returns whether a breakpoint at this item is isFlagged.
	// Two consecutive isFlagged Breakpoints are assessed an additional penalty.
	IsFlaggedBreakpoint() bool

	// IsValidBreakpoint returns whether a line can break at this item.
	// Its single argument is the preceding Item in the line.
	IsValidBreakpoint(Item) bool

	// IsBox returns whether the item is a box.
	// In general use of this method should be avoided and code should instead rely on the other polymorphic
	// methods.
	IsBox() bool

	// IsPenalty returns whether the item is a penalty.
	IsPenalty() bool
}

// NewBox creates and returns a new box item.
func NewBox(width d.Distance) Item {
	return &box{width: width}
}

type box struct {
	width d.Distance
}

func (box *box) Width() d.Distance {
	return box.width
}

func (box *box) EndOfLineWidth() d.Distance {
	return box.width
}

func (*box) Shrinkability() d.Distance {
	return 0
}

func (*box) Stretchability() d.Distance {
	return 0
}

func (*box) BreakpointPenalty() PenaltyCost {
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

func (*box) IsPenalty() bool {
	return false
}

// NewGlue creates and returns a new glue item.
func NewGlue(width d.Distance, shrinkability d.Distance, stretchability d.Distance) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: stretchability}
}

type glue struct {
	width          d.Distance
	shrinkability  d.Distance
	stretchability d.Distance
}

func (glue *glue) Width() d.Distance {
	return glue.width
}

func (glue *glue) EndOfLineWidth() d.Distance {
	return 0
}

func (glue *glue) Shrinkability() d.Distance {
	return glue.shrinkability
}

func (glue *glue) Stretchability() d.Distance {
	return glue.stretchability
}

func (*glue) BreakpointPenalty() PenaltyCost {
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

func (*glue) IsPenalty() bool {
	return false
}

type PenaltyCost int64

func (cost PenaltyCost) IsPositiveInfinite() bool {
	return cost >= PosInfBreakpointPenalty
}

func (cost PenaltyCost) IsNegativeInfinite() bool {
	return cost <= NegInfBreakpointPenalty
}

// PosInfBreakpointPenalty represents an infinitely positive penalty which makes it forbidden for the associated item
// to be a breakpoint. Any breakpoint penalty larger than this is also considered positive infinite.
const PosInfBreakpointPenalty PenaltyCost = 10000

// NegInfBreakpointPenalty represents an infinitely negative penalty which makes it mandatory for the associated item
// to be a breakpoint. Any breakpoint penalty more negative than this is also considered negative infinite.
const NegInfBreakpointPenalty PenaltyCost = -10000

// InfiniteStretchability is a stretchability constant such that the item can be stretched an arbitrary amount with
// no breakpoint penalty (or at least, no contribution to the adjustment ratio).
// Any stretchability larger than InfiniteStretchability is, itself, considered infinite.
const InfiniteStretchability d.Distance = 100000

// NewPenalty creates and returns a new penalty item.
func NewPenalty(width d.Distance, breakpointPenalty PenaltyCost, isFlagged bool) Item {
	return &penalty{width: width, breakpointPenalty: breakpointPenalty, isFlagged: isFlagged}
}

type penalty struct {
	width             d.Distance
	breakpointPenalty PenaltyCost
	isFlagged         bool
}

func (penalty *penalty) Width() d.Distance {
	return 0
}

func (penalty *penalty) EndOfLineWidth() d.Distance {
	return penalty.width
}

func (*penalty) Shrinkability() d.Distance {
	return 0
}

func (*penalty) Stretchability() d.Distance {
	return 0
}

func (penalty *penalty) BreakpointPenalty() PenaltyCost {
	return penalty.breakpointPenalty
}

func (penalty *penalty) IsFlaggedBreakpoint() bool {
	return penalty.isFlagged
}

func (penalty *penalty) IsValidBreakpoint(Item) bool {
	return !penalty.BreakpointPenalty().IsPositiveInfinite()
}

func (*penalty) IsBox() bool {
	return false
}

func (*penalty) IsPenalty() bool {
	return true
}

// ItemList is a data structure representing an ordered list of items and common operations on them.
//
// It can be used to find the total width, shrinkability and stretchability of a list of items representing a line.
// These computations incorporate various pieces of logic; for example,
//
// - a glue item at the end of a lineIndex does not contribute to any of these quantities
//
// - a penalty item only contributes if it is at the end of a lineIndex.
//
// - all items before the first box in a lineIndex are ignored
//
// In addition, the data structure is implemented such that all of these computations are fast (constant in time).
type ItemList struct {
	aggregateWidth          []d.Distance
	aggregateShrinkability  []d.Distance
	aggregateStretchability []d.Distance
	numInfStretchableItems  []int
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
		numInfStretchableItems:  itemList.numInfStretchableItems[a : b+1],
		positionToNextBoxOffset: itemList.positionToNextBoxOffset[a:b],
		items:                   itemList.items[a:b],
	}
}

// NewItemList constructs an ItemList from the provided slice of Items.
func NewItemList(items []Item) *ItemList {
	lineData := &ItemList{
		aggregateWidth:          make([]d.Distance, len(items)+1),
		aggregateShrinkability:  make([]d.Distance, len(items)+1),
		aggregateStretchability: make([]d.Distance, len(items)+1),
		numInfStretchableItems: make([]int, len(items)+1),
		positionToNextBoxOffset: make([]int, len(items)),
		items:                   items,
	}
	lineData.aggregateWidth[0] = 0
	lineData.aggregateShrinkability[0] = 0
	lineData.aggregateStretchability[0] = 0
	lineData.numInfStretchableItems[0] = 0
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
		lineData.numInfStretchableItems[position+1] = lineData.numInfStretchableItems[position]
		if item.Stretchability() >= InfiniteStretchability {
			lineData.numInfStretchableItems[position+1] += 1
		}
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
func (itemList *ItemList) Width() d.Distance {
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
func (itemList *ItemList) Shrinkability() d.Distance {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	return itemList.aggregateShrinkability[len(itemList.items)] -
		itemList.aggregateShrinkability[nextBoxIndex] -
		itemList.items[len(itemList.items)-1].Shrinkability()
}

// Stretchability returns the stretchability of a line consisting of elements of the ItemList.
func (itemList *ItemList) Stretchability() d.Distance {
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

// NumInfStretchableItems returns the number of items in the ItemList that are infinitely stretchable.
func (itemList *ItemList) NumInfStretchableItems() int {
	firstBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	return itemList.numInfStretchableItems[len(itemList.items)-1] - itemList.numInfStretchableItems[firstBoxIndex]
}