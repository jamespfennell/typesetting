package knuthplass

import (
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"github.com/jamespfennell/typesetting/pkg/knuthplass/primitives"
	"testing"
)

// Infinite stretch with negative stretch box
// 2 infinite stretches with other stretchable box
// 2 finite stretches of different stretchability

// Underfull line
// Penalty with non-zero width at the end of the line
// Invisible items at the end
// Mark penalty items as invisible if not at the end

func TestSetLine_ErrorCases(t *testing.T) {
	paramsList := []struct {
		name               string
		items              []primitives.Item
		lineLength         d.Distance
		expected           []FixedItem
		expectedLineLength d.Distance
	}{
		{
			"No boxes in the line",
			[]primitives.Item{
				primitives.NewGlue(20, 10, 10),
				primitives.NewPenalty(20, 0, false),
				primitives.NewGlue(20, 10, 10),
				primitives.NewPenalty(20, 0, false),
			},
			30,
			[]FixedItem{
				{false, 0},
				{false, 0},
				{false, 0},
				{false, 0},
			},
			0,
		},
		{
			"Underfull line",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 0),
				primitives.NewBox(20),
			},
			70,
			[]FixedItem{
				{true, 20},
				{true, 20},
				{true, 20},
			},
			60,
		},
		{
			"Very negative adjustment ratio / can't shrink line enough",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
			},
			30,
			[]FixedItem{
				{true, 20},
				{true, 10},
				{true, 20},
				{true, 10},
				{true, 20},
			},
			80,
		},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			actual, err := SetLine(primitives.NewItemList(params.items), params.lineLength)
			verifySetLineError(t, err, params.expectedLineLength)
			verifyItemsEqual(t, params.expected, actual)
		})
	}
}

func TestSetLine_NoErrorCases(t *testing.T) {
	paramsList := []struct {
		name       string
		items      []primitives.Item
		lineLength d.Distance
		expected   []FixedItem
	}{
		{
			"Single finitely stretchable item",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 20, 10),
				primitives.NewBox(20),
			},
			65,
			[]FixedItem{
				{true, 20},
				{true, 25},
				{true, 20},
			},
		},
		{
			"Perfectly fitting list of boxes",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewBox(20),
				primitives.NewBox(20),
			},
			60,
			[]FixedItem{
				{true, 20},
				{true, 20},
				{true, 20},
			},
		},
		{
			"Two equally finitely stretchable items, tie break for last sp",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 20, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 20, 10),
				primitives.NewBox(20),
			},
			111,
			[]FixedItem{
				{true, 20},
				{true, 26},
				{true, 20},
				{true, 25},
				{true, 20},
			},
		},
		{
			"Two unequally finitely stretchable items",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 20, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 12, 6),
				primitives.NewBox(20),
			},
			108,
			[]FixedItem{
				{true, 20},
				{true, 25},
				{true, 20},
				{true, 23},
				{true, 20},
			},
		},
		{
			"Single shrinkable item",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 20),
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
				primitives.NewGlue(20, 10, 20),
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 20),
				primitives.NewBox(20),
			},
			91,
			[]FixedItem{
				{true, 20},
				{true, 15},
				{true, 20},
				{true, 16},
				{true, 20},
			},
		},
		{
			"Two unequally shrinkable items",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 20),
				primitives.NewBox(20),
				primitives.NewGlue(20, 6, 12),
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
		{
			"Glue at end of line is ignored",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
			},
			60,
			[]FixedItem{
				{true, 20},
				{true, 20},
				{true, 20},
				{false, 0},
			},
		},
		{
			"Interior penalty item is invisible",
			[]primitives.Item{
				primitives.NewBox(20),
				primitives.NewGlue(20, 10, 10),
				primitives.NewPenalty(20, 0, false),
				primitives.NewBox(20),
			},
			60,
			[]FixedItem{
				{true, 20},
				{true, 20},
				{false, 0},
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

func verifySetLineError(t *testing.T, err *SetLineError, expectedLineLength d.Distance) {
	if err == nil {
		t.Errorf("Expected error but did not recieve one!")
	} else if err.ActualLineLength != expectedLineLength {
		t.Errorf("Expected line length of %d but recieved %d", expectedLineLength, err.ActualLineLength)
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
