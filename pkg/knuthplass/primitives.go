package knuthplass

import (
	"errors"
)

// TODO: rename PositiveInfinitePenalty
const PositiveInfinity int64 = 10000
const NegativeInfinity int64 = -10000
const InfiniteStretchability int64 = 100000

// what about infinite stretch?

// Item is any element in the typesetting input: box, glue or penalty
type Item interface {
	Width() int64
	EndOfLineWidth() int64
	Shrinkability() int64
	Stretchability() int64
	PenaltyCost() int64 // rename penalty
	IsFlaggedPenalty() bool
}

// NewBox creates and returns a new box item
func NewBox(width int64) Item {
	return &box{width: width}
}

// IsBox returns true iff the item is a box item
func IsBox(item Item) bool {
	_, ok := item.(*box)
	return ok
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

func (*box) PenaltyCost() int64 {
	return 0
}

func (*box) IsFlaggedPenalty() bool {
	return false
}

// NewGlue creates and returns a new glue item
func NewGlue(width int64, shrinkability int64, stretchability int64) Item {
	return &glue{width: width, shrinkability: shrinkability, stretchability: stretchability}
}

// IsGlue returns true iff the item is a glue item
func IsGlue(item Item) bool {
	_, ok := item.(*glue)
	return ok
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

func (*glue) PenaltyCost() int64 {
	return 0
}

func (*glue) IsFlaggedPenalty() bool {
	return false
}

// NewPenalty creates and returns a new penalyu item
func NewPenalty(width int64, cost int64, flagged bool) Item {
	return &penalty{width: width, cost: cost, flagged: flagged}
}

// IsPenalty returns true iff the item is a penalty item
func IsPenalty(item Item) bool {
	_, ok := item.(*penalty)
	return ok
}

type penalty struct {
	width   int64
	cost    int64
	flagged bool
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

func (penalty *penalty) PenaltyCost() int64 {
	return penalty.cost
}

func (penalty *penalty) IsFlaggedPenalty() bool {
	return penalty.flagged
}

func (penalty *penalty) IsValidBreakpoint(Item) bool {
	return penalty.PenaltyCost() < 10000
}

// IsValidBreakpoint determines whether item is a valid breakpoint
func IsValidBreakpoint(preceedingItem Item, item Item) bool {
	// It would be nice to implement this method without type casts
	// and use polymorphism instead but it's tricky because we need
	// to inspect two boxes
	switch v := item.(type) {
	case *penalty:
		return v.PenaltyCost() < PositiveInfinity
	case *glue:
		if preceedingItem == nil {
			return false
		}
		_, preceedingItemIsBox := preceedingItem.(*box)
		return preceedingItemIsBox
	default:
		return false
	}
}

// ItemList is a data structure that facilitates computing the width, shrinkability
// and stretchability of a lineIndex (i.e., a contiguous subsequence of Items) in O(1).
// It incorporates the following logic:
// - a glue item at the end of a lineIndex does not contribute to any of these quantities
// - a penalty item only contributes if it is at the end of a lineIndex.
// - all items before the first box in a lineIndex are ignored
type ItemList struct {
	aggregateWidth          []int64
	aggregateShrinkability  []int64
	aggregateStretchability []int64
	positionToNextBoxOffset []int
	items                   []Item
}

func (itemList *ItemList) Length() int {
	return len(itemList.items)
}

func (itemList *ItemList) Get(index int) Item {
	if index < 0 {
		return nil
	}
	return itemList.items[index]
}

// Slice returns a subsequence of the ItemList.
// It's analogous to sequence slicing both in terms of semantics and implementation
func (itemList *ItemList) Slice(a int, b int) *ItemList {
	return &ItemList{
		aggregateWidth:          itemList.aggregateWidth[a : b+1],
		aggregateShrinkability:  itemList.aggregateShrinkability[a : b+1],
		aggregateStretchability: itemList.aggregateStretchability[a : b+1],
		positionToNextBoxOffset: itemList.positionToNextBoxOffset[a:b],
		items:                   itemList.items[a:b],
	}
}

// NewItemList constructs an ItemList from the provided slice of Items
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
		if !IsBox(items[boxIndex]) {
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

// FirstBoxIndex returns the index of the first box in this ItemList
// or an error if the ItemList contains no box. It is equivalent to iterating
// over the list and checking IsBox for each item; however, the call is O(1).
func (itemList *ItemList) FirstBoxIndex() (int, error) {
	nextBoxIndex := itemList.positionToNextBoxOffset[0]
	if nextBoxIndex < 0 || nextBoxIndex >= len(itemList.items) {
		return -1, errors.New("this ItemList contains no boxes")
	}
	return nextBoxIndex, nil
}

// Width returns the width of a lineIndex consisting of elements of the ItemList.
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

// Shrinkability returns the shrinkability of a lineIndex consisting of elements of the ItemList.
func (itemList *ItemList) Shrinkability() int64 {
	nextBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return 0
	}
	return itemList.aggregateShrinkability[len(itemList.items)] -
		itemList.aggregateShrinkability[nextBoxIndex] -
		itemList.items[len(itemList.items)-1].Shrinkability()
}

// Stretchability returns the stretchability of a lineIndex consisting of elements of the ItemList.
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
