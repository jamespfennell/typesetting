package main

import (
	"github.com/jamespfennell/typesetting/pkg/knuthplass/primitives"
)

func main() {

	_ = []primitives.Item{
		primitives.NewBox(60),
		primitives.NewGlue(20, 7, 20),
		primitives.NewBox(60),
		primitives.NewGlue(20, 7, 20),
		primitives.NewBox(60),
		primitives.NewGlue(20, 7, 20),
		primitives.NewBox(60),
		primitives.NewGlue(0, 0, primitives.InfiniteStretchability),
		primitives.NewPenalty(0, primitives.NegInfBreakpointPenalty, false),
	}
	// expectedBreakpoints := []int{5, 8}
	// criteria := knuthplass.TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	// actualBreakpoints, err :=
	// knuthplass.CalculateBreakpoints(knuthplass.NewItemList(items), knuthplass.NewConstantLineLengths(270), criteria, false, true)

}
