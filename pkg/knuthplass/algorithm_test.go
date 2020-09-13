package knuthplass

import (
	"testing"
)

// Test forbidden breaks <- easy
// Test variable line lengths: 200, 200, 80,...
// Test setting small lines with too big a box -> no infinite loop
// No solution because force breaks make the adjustment ratio too big
// A large "happy case" test with many lines

func TestForcedBreaks(t *testing.T) {
	items := []Item{
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(50),
		NewPenalty(0, NegInfBreakpointPenalty, false),

		NewBox(60),
		NewGlue(10, 5, 5),
		NewBox(10),
		NewGlue(10, 5, 5),
		NewBox(40),
		NewPenalty(0, NegInfBreakpointPenalty, false),

		NewBox(40),
		NewGlue(0, 0, 100000),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(200), criteria, true)
	expectedBreakpoints := []int{3, 9, 12}
	testResult(t, expectedBreakpoints, result)
}

func TestBasicCase(t *testing.T) {
	items := []Item{
		NewBox(60),
		NewGlue(20, 7, 20),
		NewBox(60),
		NewGlue(20, 7, 20),
		NewBox(60),
		NewGlue(20, 7, 20), // Expected first breakpoint

		NewBox(60),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	expectedBreakpoints := []int{5, 8}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(270), criteria, true)
	testResult(t, expectedBreakpoints, result)
}

func TestConsecutiveFlaggedBreakpoint(t *testing.T) {
	// Each block has width 120, or 100 if it comes at the end of a line. It can break in the middle for free, or
	// at the end for free, but the end break will be a flagged penalty. In this test we have a line length of 340,
	// leading to an optimal line-setting of 3 lines: (3 blocks, 3 blocks, 1 block). However the two Breakpoints here
	// are flagged. By assessing a large cost for the flagged penalty, we test that the algorithm will not choose this
	// no-longer-optimal list of Breakpoints.
	block := []Item{
		NewBox(40),
		NewGlue(20, 12, 20),
		NewBox(40),
		NewPenalty(0, 0, true),
		NewGlue(20, 12, 20),
	}
	var items []Item
	for i := 0; i < 7; i++ {
		items = append(items, block...)
	}
	items = append(
		items,
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	)
	paramsList := []struct {
		name                          string
		expectedBreakpoints           []int
		consecutiveFlaggedPenaltyCost float64
	}{
		{"No consecutive flagged penalty cost", []int{13, 28, 36}, 0},
		{"Large consecutive flagged penalty cost", []int{11, 26, 36}, 20000},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			criteria := TexOptimalityCriteria{
				MaxAdjustmentRatio:            10,
				ConsecutiveFlaggedPenaltyCost: params.consecutiveFlaggedPenaltyCost,
			}
			result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(340), criteria, true)
			_ = result
			testResult(t, params.expectedBreakpoints, result)
		})
	}
}

func TestDifferentClasses(t *testing.T) {
	items := []Item{
		NewBox(80),
		NewGlue(20, 12, 20),
		NewBox(80),
		NewGlue(20, 12, 20), // Expected first breakpoint

		NewBox(40),
		NewGlue(20, 12, 21),
		NewBox(40),
		NewGlue(20, 12, 21),
		NewBox(40),
		// The following glue item is one of two candidates for the second breakpoint; the other is the final
		// penalty item. If there is no MismatchingFitnessClassCost, then the break will occur at the last item and
		// result in a very tight second line. However, the MismatchingFitnessClassCost penalizes this tight second
		// line after the loose first line, leading to the following glue item being the preferred breakpoint.
		NewGlue(20, 12, 21),

		NewBox(40),
		NewGlue(0, 0, InfiniteStretchability), // Cannot break here because of single penalty after.
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	paramsList := []struct {
		name                        string
		expectedBreakpoints         []int
		mismatchingFitnessClassCost float64
	}{
		{"No mismatching fitness class cost", []int{3, 12}, 0},
		{"Large mismatching fitness class cost", []int{3, 9, 12}, 20000},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			criteria := TexOptimalityCriteria{
				MaxAdjustmentRatio:          10,
				MismatchingFitnessClassCost: params.mismatchingFitnessClassCost,
			}
			result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(200), criteria, true)
			testResult(t, params.expectedBreakpoints, result)
		})
	}
}

func testResult(t *testing.T, expectedBreakpoints []int, result CalculateBreakpointsResult) {
	if result.Err != nil {
		t.Errorf("Solvable case marked as unsolved!")
	}
	if !listEqual(expectedBreakpoints, result.Breakpoints) {
		t.Errorf("Breakpoints not equal!")
		t.Errorf("Expected = %v != %v = actual", expectedBreakpoints, result.Breakpoints)
		result.Logger.AdjustmentRatiosTable.Print()
		result.Logger.LineDemeritsTable.Print()
		result.Logger.TotalDemeritsTable.Print()
	}
}

func listEqual(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for index, aElement := range a {
		if b[index] != aElement {
			return false
		}
	}
	return true
}
