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
// and stretchability of a line (i.e., a contiguous subsequence of Items) in O(1).
// It incorporates the following logic:
// - a glue item at the end of a line does not contribute to any of these quantities
// - a penalty item only contributes if it is at the end of a line.
// - all items before the first box in a line are ignored
type ItemList struct {
	aggregateWidth            []int64
	aggregateShrinkability    []int64
	aggregateStretchibility   []int64
	positionToNextBoxPosition map[int]int
	items                     []Item
}

func buildLineData(items []Item) *ItemList {
	lineData := &ItemList{
		aggregateWidth:            make([]int64, len(items)+1),
		aggregateShrinkability:    make([]int64, len(items)+1),
		aggregateStretchibility:   make([]int64, len(items)+1),
		positionToNextBoxPosition: make(map[int]int),
		items:                     items,
	}
	lineData.aggregateWidth[0] = 0
	lineData.aggregateShrinkability[0] = 0
	lineData.aggregateStretchibility[0] = 0
	for position, item := range items {
		lineData.aggregateWidth[position+1] =
			lineData.aggregateWidth[position] +
				item.Width()
		lineData.aggregateShrinkability[position+1] =
			lineData.aggregateShrinkability[position] +
				item.Shrinkability()
		lineData.aggregateStretchibility[position+1] =
			lineData.aggregateStretchibility[position] +
				item.Stretchability()
	}
	itemIndex := 0
	boxIndex := 0
	for boxIndex < len(items) {
		if !IsBox(items[boxIndex]) {
			boxIndex++
			continue
		}
		lineData.positionToNextBoxPosition[itemIndex] = boxIndex
		if itemIndex == boxIndex {
			boxIndex++
		}
		itemIndex++
	}
	return lineData
}

func (lineData *ItemList) getNextBoxIndex(itemIndex int) (int, error) {
	nextBoxIndex, ok := lineData.positionToNextBoxPosition[itemIndex]
	if !ok {
		return -1, errors.New("no box after this item")
	}
	return nextBoxIndex, nil
}

func (lineData *ItemList) GetWidth(previousBreakpoint int, thisBreakpoint int) int64 {
	nextBoxIndex, err := lineData.getNextBoxIndex(previousBreakpoint + 1)
	if err != nil {
		return 0
	}
	if nextBoxIndex > thisBreakpoint {
		return 0
	}
	return lineData.aggregateWidth[thisBreakpoint+1] -
		lineData.aggregateWidth[nextBoxIndex] +
		lineData.items[thisBreakpoint].EndOfLineWidth() -
		lineData.items[thisBreakpoint].Width()
}

func (lineData *ItemList) GetShrinkability(previousBreakpoint int, thisBreakpoint int) int64 {
	nextBoxIndex, err := lineData.getNextBoxIndex(previousBreakpoint + 1)
	if err != nil {
		return 0
	}
	if nextBoxIndex > thisBreakpoint {
		return 0
	}
	return lineData.aggregateShrinkability[thisBreakpoint+1] -
		lineData.aggregateShrinkability[nextBoxIndex] -
		lineData.items[thisBreakpoint].Shrinkability()
}
func (lineData *ItemList) GetStrechability(previousBreakpoint int, thisBreakpoint int) int64 {
	nextBoxIndex, err := lineData.getNextBoxIndex(previousBreakpoint + 1)
	if err != nil {
		return 0
	}
	if nextBoxIndex > thisBreakpoint {
		return 0
	}
	rawStretchability := lineData.aggregateStretchibility[thisBreakpoint+1] -
		lineData.aggregateStretchibility[nextBoxIndex] -
		lineData.items[thisBreakpoint].Stretchability()
	if rawStretchability < InfiniteStretchability {
		return rawStretchability
	}
	return InfiniteStretchability
}
