package knuthplass

import (
	"fmt"
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"github.com/jamespfennell/typesetting/pkg/knuthplass/primitives"
)

// FixedItem represents an item after its visibility and width have been set by the Knuth-Plass algorithm.
type FixedItem struct {
	// Visible denotes whether the FixedItem is visible after the line has been set.
	// This is not equivalent to the width being 0 - items with 0 width may still be visible.
	Visible bool

	// Width denotes the width of the FixedItem as determined by Knuth-Plass.
	Width d.Distance
}

// SetLineError is returned if the list of items could not be set in the given line length.
type SetLineError struct {
	// TargetLineLength is the desired with of the line.
	TargetLineLength d.Distance

	// ActualLineLength is the best width the algorithm could achieve.
	ActualLineLength d.Distance
}

// Error returns a human readable error string for a SetLineError.
func (err *SetLineError) Error() string {
	if err.IsOverfull() {
		return fmt.Sprintf("Overfull line: contents of width %s exceed line width %s ", err.TargetLineLength,
			err.ActualLineLength)
	}
	return fmt.Sprintf("Underfull line: contents of width %s do not fit line width %s ", err.TargetLineLength,
		err.ActualLineLength)
}

// IsOverfull returns true if the set line is larger than desired.
func (err *SetLineError) IsOverfull() bool {
	return err.TargetLineLength < err.ActualLineLength
}

// IsUnderfull returns true if the set line is shorter than desired.
func (err *SetLineError) IsUnderfull() bool {
	return !err.IsOverfull()
}

// SetLine sets the given ItemList to the desired line length.
// It returns the widths and visibilities of the items through the returned list of FixedItems, which is guaranteed
// to be the same length as the ItemList.
func SetLine(itemList *primitives.ItemList, lineLength d.Distance) ([]FixedItem, *SetLineError) {
	fixedItems := make([]FixedItem, itemList.Length())
	firstBoxIndex, firstBoxIndexErr := itemList.FirstBoxIndex()
	var glueAdjustments []d.Distance
	var err *SetLineError
	switch true {
	case firstBoxIndexErr != nil:
		return fixedItems, &SetLineError{TargetLineLength: lineLength, ActualLineLength: 0}
	case itemList.NumInfStretchableItems() > 0 && itemList.Width() < lineLength:
		glueAdjustments = buildInfiniteGlueAdjustments(itemList, lineLength)
	default:
		glueAdjustments, err = buildFiniteGlueAdjustments(itemList, lineLength)
	}
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		fixedItems[i].Visible = !itemList.Get(i).IsPenalty()
		fixedItems[i].Width = itemList.Get(i).Width() + glueAdjustments[i]
	}
	fixedItems[itemList.Length()-1] = buildFinalFixedItem(itemList.Get(itemList.Length() - 1))
	return fixedItems, err
}

func buildInfiniteGlueAdjustments(
	itemList *primitives.ItemList,
	lineLength d.Distance,
) []d.Distance {
	firstBoxIndex, _ := itemList.FirstBoxIndex()
	offset := make([]d.Distance, itemList.Length()-1)
	perItemToAdd := (lineLength - itemList.Width()) / d.Distance(itemList.NumInfStretchableItems())
	remainderToAdd := (lineLength - itemList.Width()) % d.Distance(itemList.NumInfStretchableItems())
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		if !itemList.Get(i).Stretchability().IsInfinite() {
			continue
		}
		offset[i] += perItemToAdd
		if remainderToAdd > 0 {
			offset[i]++
			remainderToAdd--
		}
	}
	return offset
}

func buildFiniteGlueAdjustments(
	itemList *primitives.ItemList,
	lineLength d.Distance,
) ([]d.Distance, *SetLineError) {
	adjustmentRatio := itemList.CalculateAdjustmentRatio(lineLength)
	if adjustmentRatio.LessThan(d.MinusOneRatio) {
		adjustmentRatio = d.MinusOneRatio
	}
	var scalingPropertyGetter func(primitives.Item) d.Distance
	if adjustmentRatio.LessThan(d.ZeroRatio) {
		scalingPropertyGetter = func(item primitives.Item) d.Distance { return item.Shrinkability() }
	} else {
		scalingPropertyGetter = func(item primitives.Item) d.Distance { return item.Stretchability().FiniteValue() }
	}
	// We ignore the error because it's handled upstream
	firstBoxIndex, _ := itemList.FirstBoxIndex()
	offset := make([]d.Distance, itemList.Length()-1)
	var totalOffset d.Distance
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		if scalingPropertyGetter(itemList.Get(i)) == 0 {
			continue
		}
		// Note that the result will be at most what it should be, and potentially 1 less.
		offset[i] = (scalingPropertyGetter(itemList.Get(i)) * adjustmentRatio.Num) / adjustmentRatio.Den
		totalOffset += offset[i]
	}
	// Missing is non-zero only due to round off errors. Its magnitude will be at most the number of items
	// with non-zero scaling, and so adding or subtracting one to these items will yield the right length.
	missing := lineLength - (itemList.Width() + totalOffset)
	// In the case when the adjustment ratio is less than or equal to -1, we can't perform any more adjustments
	// as that would make the shrinkability more than the max.
	if !adjustmentRatio.LessThanEqual(d.MinusOneRatio) {
		for i := firstBoxIndex; i < itemList.Length()-1; i++ {
			if missing == 0 {
				break
			}
			if scalingPropertyGetter(itemList.Get(i)) == 0 {
				continue
			}
			if missing > 0 {
				offset[i]++
				missing--
			} else {
				offset[i]--
				missing++
			}
		}
	}
	var err *SetLineError
	if missing != 0 {
		err = &SetLineError{TargetLineLength: lineLength, ActualLineLength: lineLength - missing}
	}
	return offset, err
}

func buildFinalFixedItem(lastItem primitives.Item) FixedItem {
	return FixedItem{
		Visible: lastItem.IsPenalty() || lastItem.IsBox(),
		Width:   lastItem.EndOfLineWidth(),
	}
}
