package knuthplass

import (
	"fmt"
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"github.com/jamespfennell/typesetting/pkg/knuthplass/primitives"
)

type FixedItem struct {
	Visible bool
	Width   d.Distance
}

type SetLineError struct {
	TargetLineLength d.Distance
	ActualLineLength d.Distance
}

func (err *SetLineError) Error() string {
	if err.IsOverfull() {
		return fmt.Sprintf("Overfull line: contents of width %s exceed line width %s ", err.TargetLineLength,
			err.ActualLineLength)
	}
	return fmt.Sprintf("Underfull line: contents of width %s do not fit line width %s ", err.TargetLineLength,
		err.ActualLineLength)
}

func (err *SetLineError) IsOverfull() bool {
	return err.TargetLineLength < err.ActualLineLength
}

func (err *SetLineError) IsUnderfull() bool {
	return !err.IsOverfull()
}

func SetLine(itemList *primitives.ItemList, lineLength d.Distance) ([]FixedItem, *SetLineError) {
	fixedItems := make([]FixedItem, itemList.Length())
	firstBoxIndex, firstBoxIndexErr := itemList.FirstBoxIndex()
	if firstBoxIndexErr != nil {
		return fixedItems, &SetLineError{TargetLineLength: lineLength, ActualLineLength: 0}
	}
	offset := make([]d.Distance, itemList.Length())
	adjustmentRatio := calculateAdjustmentRatio(
		itemList.Width(), itemList.Shrinkability(), itemList.Stretchability(), lineLength)
	var err *SetLineError
	if adjustmentRatio.LessThan(d.MinusOneRatio) {
		err = &SetLineError{TargetLineLength: lineLength, ActualLineLength: itemList.Width() - itemList.Stretchability()}
		adjustmentRatio = d.MinusOneRatio
	}
	offset = buildAdjustmentSummands(itemList, lineLength, adjustmentRatio)
	for i := firstBoxIndex; i < itemList.Length(); i++ {
		fixedItems[i] = FixedItem{Visible: true, Width: itemList.Get(i).Width() + offset[i]}
	}
	lastItem := itemList.Get(itemList.Length() - 1)
	lastItemWidthOffset := lastItem.EndOfLineWidth() - lastItem.Width()
	if lastItemWidthOffset != 0 {
		fixedItems[itemList.Length()-1].Width += lastItemWidthOffset
		// Assumption: the last item having an offset means it's a glue item to be discarded.
		fixedItems[itemList.Length()-1].Visible = false
	}
	return fixedItems, err
}

func getShrinkability(item primitives.Item) d.Distance {
	return item.Shrinkability()
}

func getStretchability(item primitives.Item) d.Distance {
	return item.Stretchability()
}

func buildAdjustmentSummands(
	itemList *primitives.ItemList,
	lineLength d.Distance,
	adjustmentRatio d.Ratio,
) []d.Distance {
	var scalingPropertyGetter func(primitives.Item) d.Distance
	if adjustmentRatio.LessThan(d.ZeroRatio) {
		scalingPropertyGetter = getShrinkability
	} else {
		scalingPropertyGetter = getStretchability
	}
	// We ignore the error because it's handled upstream
	firstBoxIndex, _ := itemList.FirstBoxIndex()
	offset := make([]d.Distance, itemList.Length())
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
	// with non-zero scaling
	missing := lineLength - (itemList.Width() + totalOffset)
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		if missing == 0 {
			break
		}
		// In no case will we shrink an item more than its shrinkability allows.
		// TODO: I bet there's an edge case where the line length is not what it should be because of this
		if -scalingPropertyGetter(itemList.Get(i)) == offset[i] {
			continue
		}
		if missing > 0 {
			offset[i] += 1
			missing -= 1
		} else {
			offset[i] -= 1
			missing += 1
		}
	}
	return offset
}
