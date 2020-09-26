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
	firstBoxIndex, err := itemList.FirstBoxIndex()
	if err != nil {
		return fixedItems, &SetLineError{TargetLineLength: lineLength, ActualLineLength: 0}
	}
	offset := make([]d.Distance, itemList.Length())
	adjustmentRatio := calculateAdjustmentRatio(
		itemList.Width(), itemList.Shrinkability(), itemList.Stretchability(), lineLength)
	if adjustmentRatio.LessThan(d.ZeroRatio) {
		var totalOffset d.Distance
		for i := firstBoxIndex; i < itemList.Length(); i++ {
			if itemList.Get(i).Shrinkability() == 0 {
				continue
			}
			// Will in general under shrink due to the round off error
			offset[i] = (itemList.Get(i).Shrinkability() * adjustmentRatio.Num) / adjustmentRatio.Den
			totalOffset += offset[i]
		}
		// Missing is always positive because we haven't shrunk the itemList enough yet
		// Missing is non-zero only due to round off errors. It will be at most the number of non-zero shrinkable
		// items
		missing := itemList.Width() + totalOffset - lineLength
		for i := firstBoxIndex; i < itemList.Length(); i++ {
			if missing == 0 {
				break
			}
			if itemList.Get(i).Shrinkability() == 0 {
				continue
			}
			offset[i] -= 1
			missing -= 1
		}

	}
	for i := firstBoxIndex; i < itemList.Length(); i++ {
		fixedItems[i] = FixedItem{Visible: true, Width: itemList.Get(i).Width() + offset[i]}
	}
	return fixedItems, nil
}
