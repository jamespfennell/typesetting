package main

import (
	"github.com/jamespfennell/typesetting/pkg/knuthplass"
)

func main() {

	items := []knuthplass.Item{
		knuthplass.NewBox(60),
		knuthplass.NewGlue(20, 7, 20),
		knuthplass.NewBox(60),
		knuthplass.NewGlue(20, 7, 20),
		knuthplass.NewBox(60),
		knuthplass.NewGlue(20, 7, 20),
		knuthplass.NewBox(60),
		knuthplass.NewGlue(0, 0, knuthplass.InfiniteStretchability),
		knuthplass.NewPenalty(0, knuthplass.NegInfBreakpointPenalty, false),
	}
	// expectedBreakpoints := []int{5, 8}
	criteria := knuthplass.TexOptimalityCriteria{MaxAdjustmentRatio: 200000}
	// actualBreakpoints, err :=
	knuthplass.CalculateBreakpoints(knuthplass.NewItemList(items), knuthplass.NewConstantLineLengths(270), criteria, true)

}
