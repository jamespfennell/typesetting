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
	Stretchability() Stretchability

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

// Stretchability represents the stretchability of an Item.
// An item can either have a finite or infinite stretchability.
type Stretchability struct {
	numInfComponents int
	value            d.Distance
}

// Add adds two Stretchability types together.
func (stretchability Stretchability) Add(rhs Stretchability) Stretchability {
	return Stretchability{
		numInfComponents: stretchability.numInfComponents + rhs.numInfComponents,
		value:            stretchability.value + rhs.value,
	}
}

// Subtract subtracts one Stretchability type from another.
func (stretchability Stretchability) Subtract(rhs Stretchability) Stretchability {
	return Stretchability{
		numInfComponents: stretchability.numInfComponents - rhs.numInfComponents,
		value:            stretchability.value - rhs.value,
	}
}

// IsInfinite tells whether the Stretchability is infinite.
func (stretchability Stretchability) IsInfinite() bool {
	return stretchability.numInfComponents > 0
}

// FiniteValue returns the value of the Stretchability if it is finite.
// If the Stretchability is infinitely stretchable then the result of this function is undefined.
func (stretchability Stretchability) FiniteValue() d.Distance {
	return stretchability.value
}

// baseItem makes the code less verbose by providing defaults for all the methods. We embed it in each implementation
// of the Item interface and then only have to "overload" the non-default methods.
type baseItem struct{}

func (baseItem) Width() d.Distance {
	return 0
}

func (baseItem) EndOfLineWidth() d.Distance {
	return 0
}

func (baseItem) Shrinkability() d.Distance {
	return 0
}

func (baseItem) Stretchability() Stretchability {
	return Stretchability{}
}

func (baseItem) BreakpointPenalty() PenaltyCost {
	return 0
}

func (baseItem) IsFlaggedBreakpoint() bool {
	return false
}

func (baseItem) IsValidBreakpoint(Item) bool {
	return false
}

func (baseItem) IsBox() bool {
	return false
}

func (baseItem) IsPenalty() bool {
	return false
}

// NewBox creates and returns a new box item.
func NewBox(width d.Distance) Item {
	return &box{width: width}
}

type box struct {
	baseItem
	width d.Distance
}

func (box *box) Width() d.Distance {
	return box.width
}

func (box *box) EndOfLineWidth() d.Distance {
	return box.width
}

func (*box) IsBox() bool {
	return true
}

// NewGlue creates and returns a new glue item.
func NewGlue(width d.Distance, shrinkability d.Distance, stretchability d.Distance) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: Stretchability{value: stretchability}}
}

// NewInfStretchGlue creates and returns a new glue item that is infinitely stretchable.
func NewInfStretchGlue(width d.Distance, shrinkability d.Distance) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: Stretchability{numInfComponents: 1}}
}

type glue struct {
	baseItem
	width          d.Distance
	shrinkability  d.Distance
	stretchability Stretchability
}

func (glue *glue) Width() d.Distance {
	return glue.width
}

func (glue *glue) Shrinkability() d.Distance {
	return glue.shrinkability
}

func (glue *glue) Stretchability() Stretchability {
	return glue.stretchability
}

func (*glue) IsValidBreakpoint(precedingItem Item) bool {
	if precedingItem == nil {
		return false
	}
	return precedingItem.IsBox()
}

// PenaltyCost represents the cost of breaking at a penalty item.
type PenaltyCost int64

// IsPositiveInfinite returns whether this PenaltyCost is plus infinity.
// In this case the associated breakpoint is forbidden.
func (cost PenaltyCost) IsPositiveInfinite() bool {
	return cost >= PosInfBreakpointPenalty
}

// IsNegativeInfinite returns whether this PenaltyCost is negative infinity.
// In this case the associated breakpoint is mandatory.
func (cost PenaltyCost) IsNegativeInfinite() bool {
	return cost <= NegInfBreakpointPenalty
}

// PosInfBreakpointPenalty represents an infinitely positive penalty which makes it forbidden for the associated item
// to be a breakpoint. Any breakpoint penalty larger than this is also considered positive infinite.
const PosInfBreakpointPenalty PenaltyCost = 10000

// NegInfBreakpointPenalty represents an infinitely negative penalty which makes it mandatory for the associated item
// to be a breakpoint. Any breakpoint penalty more negative than this is also considered negative infinite.
const NegInfBreakpointPenalty PenaltyCost = -10000

// NewPenalty creates and returns a new penalty item.
func NewPenalty(width d.Distance, breakpointPenalty PenaltyCost, isFlagged bool) Item {
	return &penalty{width: width, breakpointPenalty: breakpointPenalty, isFlagged: isFlagged}
}

type penalty struct {
	baseItem
	width             d.Distance
	breakpointPenalty PenaltyCost
	isFlagged         bool
}

func (penalty *penalty) EndOfLineWidth() d.Distance {
	return penalty.width
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
	aggregateStretchability []Stretchability
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
		aggregateWidth:          make([]d.Distance, len(items)+1),
		aggregateShrinkability:  make([]d.Distance, len(items)+1),
		aggregateStretchability: make([]Stretchability, len(items)+1),
		positionToNextBoxOffset: make([]int, len(items)),
		items:                   items,
	}
	lineData.aggregateWidth[0] = 0
	lineData.aggregateShrinkability[0] = 0
	lineData.aggregateStretchability[0] = Stretchability{}
	for position, item := range items {
		lineData.aggregateWidth[position+1] =
			lineData.aggregateWidth[position] +
				item.Width()
		lineData.aggregateShrinkability[position+1] =
			lineData.aggregateShrinkability[position] +
				item.Shrinkability()
		lineData.aggregateStretchability[position+1] = lineData.aggregateStretchability[position].Add(item.Stretchability())
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
func (itemList *ItemList) Stretchability() Stretchability {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return Stretchability{}
	}
	return itemList.aggregateStretchability[len(itemList.items)].Subtract(
		itemList.aggregateStretchability[nextBoxIndex]).Subtract(
		itemList.items[len(itemList.items)-1].Stretchability())
}

// NumInfStretchableItems returns the number of items in the ItemList that are infinitely stretchable.
func (itemList *ItemList) NumInfStretchableItems() int {
	return itemList.Stretchability().numInfComponents
}

// CalculateAdjustmentRatio returns the adjustment ratio when trying to set the given ItemList to a target line width.
func (itemList *ItemList) CalculateAdjustmentRatio(targetLineWidth d.Distance) d.Ratio {
	widthDifference := targetLineWidth - itemList.Width()
	switch {
	case widthDifference < 0:
		return d.Ratio{Num: widthDifference, Den: itemList.Shrinkability()}
	case widthDifference > 0:
		if itemList.Stretchability().IsInfinite() {
			return d.ZeroRatio
		}
		return d.Ratio{Num: widthDifference, Den: itemList.Stretchability().FiniteValue()}
	default:
		return d.ZeroRatio
	}
}
