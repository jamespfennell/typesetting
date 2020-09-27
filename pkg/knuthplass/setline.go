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
	adjustmentRatio := calculateAdjustmentRatio(
		itemList.Width(), itemList.Shrinkability(), itemList.Stretchability(), lineLength)

	var err *SetLineError
	// If the adjustment ratio is really small, mark an error.
	if adjustmentRatio.LessThan(d.MinusOneRatio) {
		err = &SetLineError{TargetLineLength: lineLength, ActualLineLength: itemList.Width() - itemList.Stretchability()}
		adjustmentRatio = d.MinusOneRatio
	}
	// If the adjustment ratio is positive and the line can't be stretched at all, mark an error.
	if !adjustmentRatio.LessThanEqual(d.ZeroRatio) && itemList.Stretchability() == 0 {
		err = &SetLineError{TargetLineLength: lineLength, ActualLineLength: itemList.Width()}
	}

	glueAdjustments := buildGlueAdjustments(itemList, lineLength, adjustmentRatio)
	for i := firstBoxIndex; i < itemList.Length(); i++ {
		fixedItems[i].Visible = !itemList.Get(i).IsPenalty()
		fixedItems[i].Width = itemList.Get(i).Width() + glueAdjustments[i]
	}
	lastItem := itemList.Get(itemList.Length() - 1)
	lastItemWidthOffset := lastItem.EndOfLineWidth() - lastItem.Width()
	if lastItemWidthOffset != 0 {
		fixedItems[itemList.Length()-1].Width += lastItemWidthOffset
		// TODO: this assumption is false as we will discover with a test :)
		// Assumption: the last item having a non-zero summand means it's a glue item to be discarded.
		fixedItems[itemList.Length()-1].Visible = false
	}
	return fixedItems, err
}

func buildGlueAdjustments(
	itemList *primitives.ItemList,
	lineLength d.Distance,
	adjustmentRatio d.Ratio,
) []d.Distance {
	var scalingPropertyGetter func(primitives.Item) d.Distance
	if adjustmentRatio.LessThan(d.ZeroRatio) {
		scalingPropertyGetter = func(item primitives.Item) d.Distance {return item.Shrinkability()}
	} else {
		scalingPropertyGetter = func(item primitives.Item) d.Distance {return item.Stretchability()}
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
