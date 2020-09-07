package knuthplass

import (
	"fmt"
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
		NewPenalty(0, NegativeInfinity, false),
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
