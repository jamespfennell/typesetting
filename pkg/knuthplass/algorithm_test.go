package knuthplass

import (
	"fmt"
	"reflect"
	"testing"
)

func TestForcedBreaks(t *testing.T) {
	items := []Item{
		NewBox(100),
		NewGlue(10, 5, 5),
		NewBox(50),
		NewPenalty(0, NegativeInfinity, false),
		NewBox(60),
		NewGlue(10, 5, 5),
		NewBox(10),
		NewGlue(10, 5, 5),
		NewBox(40),
		NewPenalty(0, NegativeInfinity, false),
		NewBox(40),
		NewGlue(0, 0, 100000),
		NewPenalty(0, NegativeInfinity, false),
	}
	criteria := TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	actualBreakpoints, err := KnuthPlassAlgorithm(items, []int64{}, 200, criteria)
	if err != nil {
		t.Errorf("Solvable case marked as unsolved!")
	}
	expectedBreakpoints := []int64{
		3,
		9,
		12,
	}
	if !reflect.DeepEqual(expectedBreakpoints, actualBreakpoints) {
		t.Errorf("Results not equal!")
		fmt.Println("Actual breakpoints", actualBreakpoints)

	}
}
