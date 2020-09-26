package knuthplass

import (
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
	return "Item could not be set"
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
		return nil, &SetLineError{TargetLineLength: lineLength, ActualLineLength: 0}
	}
	for i := 0; i < firstBoxIndex; i++ {
		fixedItems[i] = FixedItem{Visible: false, Width: 0}
	}
	for i := firstBoxIndex; i < itemList.Length(); i++ {
		fixedItems[i] = FixedItem{Visible: true, Width: itemList.Get(i).Width()}
	}
	return nil, nil
}
