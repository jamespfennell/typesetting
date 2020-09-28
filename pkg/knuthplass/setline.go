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
	adjustmentRatio := itemList.CalculateAdjustmentRatio(lineLength)

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

	glueAdjustments := buildFiniteGlueAdjustments(itemList, adjustmentRatio)
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		fixedItems[i].Visible = !itemList.Get(i).IsPenalty()
		fixedItems[i].Width = itemList.Get(i).Width() + glueAdjustments[i]
	}
	fixedItems[itemList.Length()-1] = buildFinalFixedItem(itemList.Get(itemList.Length() - 1))
	return fixedItems, err
}

func buildFiniteGlueAdjustments(
	itemList *primitives.ItemList,
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
	// with non-zero scaling
	if adjustmentRatio.LessThanEqual(d.MinusOneRatio) {
		return offset
	}
	missing := adjustmentRatio.Num - totalOffset
	for i := firstBoxIndex; i < itemList.Length()-1; i++ {
		if missing == 0 {
			break
		}
		if scalingPropertyGetter(itemList.Get(i)) == 0 {
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

func buildFinalFixedItem(lastItem primitives.Item) FixedItem {
	return FixedItem{
		Visible: lastItem.IsPenalty() || lastItem.IsBox(),
		Width: lastItem.EndOfLineWidth(),
	}
}