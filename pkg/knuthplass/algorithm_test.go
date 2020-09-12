package knuthplass

import (
	"fmt"
	"testing"
)

// Test forbidden breaks

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
	actualBreakpoints, err := KnuthPlassAlgorithm(NewItemList(items), NewConstantLineLengths(200), criteria)
	if err != nil {
		t.Errorf("Solvable case marked as unsolved!")
	}
	expectedBreakpoints := []int{
		3,
		9,
		12,
	}
	if !listEqual(expectedBreakpoints, actualBreakpoints) {
		t.Errorf("Results not equal!")
		fmt.Println("Actual breakpoints", actualBreakpoints)
	}
}

func TestBasicCase(t *testing.T) {
	items := []Item{
		NewBox(60),
		NewGlue(20, 7, 20),
		NewBox(60),
		NewGlue(20, 7, 20),
		NewBox(60),
		NewGlue(20, 7, 20),
		NewBox(60),
		NewGlue(0, 0, InfiniteStretchability),
		NewPenalty(0, NegInfBreakpointPenalty, false),
	}
	expectedBreakpoints := []int{5, 8}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	actualBreakpoints, err := KnuthPlassAlgorithm(NewItemList(items), NewConstantLineLengths(270), criteria)
	if err != nil {
		t.Errorf("Solvable case marked as unsolved!")
	}
	if !listEqual(expectedBreakpoints, actualBreakpoints) {
		fmt.Println("Actual breakpoints", actualBreakpoints)
		t.Errorf("Results not equal!")
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
		{"Large mismatching fitness class cost", []int{3, 9, 12}, 0},
	}
	for _, params := range paramsList {
		t.Run(params.name, func(t *testing.T) {
			criteria := TexOptimalityCriteria{
				MaxAdjustmentRatio:          10,
				MismatchingFitnessClassCost: params.mismatchingFitnessClassCost,
			}
			actualBreakpoints, err := KnuthPlassAlgorithm(NewItemList(items), NewConstantLineLengths(200), criteria)
			if err != nil {
				t.Errorf("Solvable case marked as unsolved!")
			}
			if !listEqual(params.expectedBreakpoints, actualBreakpoints) {
				fmt.Println("Expected breakpoints", params.expectedBreakpoints)
				fmt.Println("Actual breakpoints", actualBreakpoints)
				t.Errorf("Results not equal!")
			}
		})
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
