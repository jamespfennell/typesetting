package knuthplass

import (
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"github.com/jamespfennell/typesetting/pkg/knuthplass/primitives"
	"testing"
)

// Infinite stretch with negative stretch box
// 2 infinite stretches with other stretchable box
// 2 finite stretches of different stretchability
// Overfull line
// Underfull line
// Penalty with non-zero width at the end of the line
// Stretchable item at the end of the line should be ignored
// Widths fit exactly
// Invisible items at the end

func TestNoItemsInList(t *testing.T) {
	list := primitives.NewItemList([]primitives.Item{
		primitives.NewGlue(20, 10, 10),
		primitives.NewPenalty(20, 0, false),
		primitives.NewGlue(20, 10, 10),
		primitives.NewPenalty(20, 0, false),
	})
	expected := []FixedItem{
		{false, 0},
		{false, 0},
		{false, 0},
		{false, 0},
	}
	actual, err := SetLine(list, 30)
	verifySetLineError(t, err, true)
	verifyItemsEqual(t, expected, actual)
}

func TestNoErrorCases(t *testing.T) {
	paramsList := []struct {
		name       string
		items      []primitives.Item
		lineLength d.Distance
		expected   []FixedItem
	}{
		{
			"Single shrinkable item",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
			},
			55,
			[]FixedItem{
				{true, 20},
				{true, 15},
				{true, 20},
			},
		},
		{
			"Two equally shrinkable items, tie break for last sp",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
			},
			71,
			[]FixedItem{
				{true, 20},
				{true, 5},
				{true, 20},
				{true, 6},
				{true, 20},
			},
		},
		{
			"Two unequally shrinkable items",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 6, 10),
				primitives.NewBox(20),
			},
			92,
			[]FixedItem{
				{true, 20},
				{true, 15},
				{true, 20},
				{true, 17},
				{true, 20},
			},
		},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			actual, err := SetLine(primitives.NewItemList(params.items), params.lineLength)
			verifyNoSetLineError(t, err)
			verifyItemsEqual(t, params.expected, actual)
		})
	}
}

func verifySetLineError(t *testing.T, err *SetLineError, expectUnderfull bool) {
	if err == nil {
		t.Errorf("Expected error but did not recieve one!")
	} else if err.IsUnderfull() != expectUnderfull {
		if expectUnderfull {
			t.Errorf("Expected underfull error but recieved overfull error")
		} else {
			t.Errorf("Expected overfull error but recieved underfull error")
		}
	}
}

func verifyNoSetLineError(t *testing.T, err *SetLineError) {
	if err != nil {
		t.Errorf("Expected no error but recieveed one!")
	}
}

func verifyItemsEqual(t *testing.T, expected, actual []FixedItem) {
	if !fixedItemsEqual(expected, actual) {
		t.Errorf("Fixed items not equal!\nExpected: %v\nActual:   %v", expected, actual)
	}
}

func fixedItemsEqual(a, b []FixedItem) bool {
	if len(a) != len(b) {
		return false
	}
	for i, _ := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
