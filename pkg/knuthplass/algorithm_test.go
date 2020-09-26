package knuthplass

import (
	"fmt"
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"testing"
)

// We're missing a large "happy case" test with many lines, kind of hard to generate though.
// Also a test for the looseness !=0 case. Probably would be good to re-use the test described above.

func TestVariableLineLengthsBreakpoints(t *testing.T) {
	lineLengths := NewVariableLineLengths([]d.Distance{100, 200, 150}, 50)
	items := []Item{
		NewBox(50),
		NewGlue(10, 5, 5),
		NewBox(40),
		NewGlue(10, 5, 5),

		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(90),
		NewGlue(10, 5, 5),

		NewBox(50),
		NewGlue(10, 5, 5),
		NewBox(90),
		NewGlue(10, 5, 5),

		NewBox(20),
		NewGlue(10, 5, 5),
		NewBox(20),
		NewGlue(10, 5, 5),

		NewBox(20),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{10, 1}}
	result := CalculateBreakpoints(NewItemList(items), lineLengths, criteria, false, true)
	verifyNoError(t, result)
	verifyBreakpoints(t, []int{3, 7, 11, 15, 18}, result)
}

func TestBestIllegalCase(t *testing.T) {
	items := []Item{
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{10, 1}}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(150), criteria, true, true)
	verifyError(t, []int{3}, result)
	verifyBreakpoints(t, []int{3, 6}, result)
}

func TestBoxTooBig(t *testing.T) {
	items := []Item{
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{10, 1}}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(150), criteria, false, true)
	verifyError(t, []int{3}, result)
}

func TestDeterministicFinalNodeSelection(t *testing.T) {
	// This is a highly contrived test case. It is designed to that there are two optimal ways, on different lines
	// to set the paragraph. The test is to ensure that the algorithm deterministically chooses []int{3, 6} instead
	// of []int{6}.
	items := []Item{
		NewBox(30),
		NewGlue(0, 0, InfiniteStretchability),
		NewBox(30),
		NewPenalty(0, -1, false),
		NewBox(30),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{10, 1}}
	result := CalculateBreakpoints(NewItemList(items), NewVariableLineLengths([]d.Distance{90, 30}, 0), criteria, false, true)
	verifyNoError(t, result)
	verifyBreakpoints(t, []int{3, 6}, result)
}

func TestAdjustmentRatioTooBig(t *testing.T) {
	items := []Item{
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(100),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{1, 1}}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(230), criteria, false, true)
	verifyError(t, []int{5}, result)
}

func TestForbiddenBreaks(t *testing.T) {
	paramsList := []struct {
		name                string
		penalty             int64
		expectedBreakpoints []int
	}{
		{"No penalty", 0, []int{5, 9}},
		{"Infinite penalty", PosInfBreakpointPenalty, []int{3, 9}},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			items := []Item{
				NewBox(100),
				NewGlue(10, 5, 5),
				NewBox(10),
				NewGlue(10, 5, 5), // Optimal breakpoint, if penalty=+Inf

				NewBox(10),
				NewPenalty(0, params.penalty, false), // Optimal breakpoint, if penalty=0

				NewGlue(10, 5, 5),
				NewBox(10),
				NewGlue(0, 0, 100000),
				NewPenalty(0, NegInfBreakpointPenalty, false),
			}
			criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{10, 1}}
			result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(140), criteria, false, true)
			verifyNoError(t, result)
			verifyBreakpoints(t, params.expectedBreakpoints, result)
		})
	}
}

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
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{200000, 1}}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(200), criteria, false, true)
	expectedBreakpoints := []int{3, 9, 12}
	verifyNoError(t, result)
	verifyBreakpoints(t, expectedBreakpoints, result)
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
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: adjustmentRatio{200000,1}}
	result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(270), criteria, false, true)
	verifyNoError(t, result)
	verifyBreakpoints(t, expectedBreakpoints, result)
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
				MaxAdjustmentRatio:            adjustmentRatio{10, 1},
				ConsecutiveFlaggedPenaltyCost: params.consecutiveFlaggedPenaltyCost,
			}
			result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(340), criteria, false, true)
			_ = result
			verifyNoError(t, result)
			verifyBreakpoints(t, params.expectedBreakpoints, result)
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
				MaxAdjustmentRatio:          adjustmentRatio{10, 1},
				MismatchingFitnessClassCost: params.mismatchingFitnessClassCost,
			}
			result := CalculateBreakpoints(NewItemList(items), NewConstantLineLengths(200), criteria, false, true)
			verifyNoError(t, result)
			verifyBreakpoints(t, params.expectedBreakpoints, result)
		})
	}
}

func verifyBreakpoints(t *testing.T, expectedBreakpoints []int, result CalculateBreakpointsResult) {
	if !listEqual(expectedBreakpoints, result.Breakpoints) {
		t.Errorf("Breakpoints not equal!")
		t.Errorf("Expected = %v != %v = actual", expectedBreakpoints, result.Breakpoints)
		fmt.Println(result.Logger.AdjustmentRatiosTable.String())
		fmt.Println(result.Logger.LineDemeritsTable.String())
		fmt.Println(result.Logger.TotalDemeritsTable.String())
	}
}

func verifyNoError(t *testing.T, result CalculateBreakpointsResult) {
	if result.Err != nil {
		t.Errorf("Solvable case marked as unsolved!")
	}
}

func verifyError(t *testing.T, expectedProblematicIndices []int, result CalculateBreakpointsResult) {
	if result.Err == nil {
		t.Fatal("Expected error, received none")
	}
	if !listEqual(result.Err.ProblematicItemIndices, expectedProblematicIndices) {
		t.Errorf("Probelematic item indices not equal!")
		t.Errorf("Expected = %v != %v = actual", expectedProblematicIndices, result.Err.ProblematicItemIndices)
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
