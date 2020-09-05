package knuthplass

import "testing"

func TestIsValidBreakpoint(t *testing.T) {
	paramsList := []struct {
		preceedingItem    Item
		item              Item
		isValidBreakpoint bool
	}{
		{nil, &box{}, false},
		{&box{}, &box{}, false},
		{&glue{}, &box{}, false},
		{&penalty{}, &box{}, false},
		{nil, &glue{}, false},
		{&box{}, &glue{}, true},
		{&glue{}, &glue{}, false},
		{&penalty{}, &glue{}, false},
		{nil, &penalty{}, true},
		{&box{}, &penalty{}, true},
		{&glue{}, &penalty{}, true},
		{&penalty{}, &penalty{}, true},
		{nil, &penalty{cost: 20000}, false},
		{&box{}, &penalty{cost: 20000}, false},
		{&glue{}, &penalty{cost: 20000}, false},
		{&penalty{}, &penalty{cost: 20000}, false},
	}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			if IsValidBreakpoint(params.preceedingItem, params.item) != params.isValidBreakpoint {
				t.Errorf(
					"IsValidBreakpoint(%T, %T) != %t",
					params.preceedingItem,
					params.item,
					params.isValidBreakpoint,
				)
			}
		})
	}
}

func TestLineData(t *testing.T) {
	items := []Item{
		NewBox(1),
		NewGlue(2, 10, 100),
		NewBox(3),
		NewGlue(4, 20, 200),
		NewBox(5),
		NewGlue(6, 30, 300),
		NewPenalty(7, 0, false),
		NewGlue(8, 40, InfiniteStretchability),
		NewGlue(9, 50, 500),
		NewBox(10),
		NewGlue(11, 50, 600),
	}
	paramsList := []struct {
		previousBreakpoint     int
		thisBreakpoint         int
		expectedWidth          int64
		expectedShrinkability  int64
		expectedStretchability int64
	}{
		{-1, 2, 1 + 2 + 3, 10, 100},             // First line is correct
		{-1, 3, 1 + 2 + 3, 10, 100},             // First line is correct + end glue ignored
		{0, 2, 3, 0, 0},                         // First glue ignored
		{2, 6, 5 + 6 + 7, 30, 300},              // Penalty width counted
		{4, 9, 10, 0, 0},                        // Glue + penalty at start ignored
		{4, 8, 0, 0, 0},                         // No box on the line
		{9, 10, 0, 0, 0},                        // No box on the line (end of paragraph)
		{1, 9, 45, 140, InfiniteStretchability}, // Happy path
	}

	lineData := buildLineData(items)
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			actualWidth := lineData.GetWidth(params.previousBreakpoint,
				params.thisBreakpoint)
			if params.expectedWidth != actualWidth {
				t.Errorf(
					"lineData.GetWidth(%d, %d) = %d != %d",
					params.previousBreakpoint,
					params.thisBreakpoint,
					actualWidth,
					params.expectedWidth,
				)
			}
		})
	}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			actualShrinkability := lineData.GetShrinkability(params.previousBreakpoint,
				params.thisBreakpoint)
			if params.expectedShrinkability != actualShrinkability {
				t.Errorf(
					"lineData.GetShrinkability(%d, %d) = %d != %d",
					params.previousBreakpoint,
					params.thisBreakpoint,
					actualShrinkability,
					params.expectedShrinkability,
				)
			}
		})
	}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			actualStretchability := lineData.GetStrechability(params.previousBreakpoint,
				params.thisBreakpoint)
			if params.expectedStretchability != actualStretchability {
				t.Errorf(
					"lineData.GetStretchability(%d, %d) = %d != %d",
					params.previousBreakpoint,
					params.thisBreakpoint,
					actualStretchability,
					params.expectedStretchability,
				)
			}
		})
	}
}
