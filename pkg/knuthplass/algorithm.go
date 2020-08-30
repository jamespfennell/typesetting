package knuthplass

import (
	"fmt"
)

func KnuthPlassAlgorithm(
	items []Item,
	startingLineLengths []int64,
	subsequentLineLengths int64,
	criteria OptimalityCriteria,
) ([]int, error) {

	lineData := buildLineData(items)

	// Set of nodes
	activeNodes := make(map[node]bool)
	newActiveNodes := make(map[node]bool)
	firstNode := node{position: -1, line: 0, fitnessClass: 1}
	activeNodes[firstNode] = true

	// Data about the nodes. We don't include this data on the node object itself
	// as that would interfere with hashing
	nodeToPrevious := make(map[node]node)
	nodeToMinDemerits := make(map[node]float64)

	for position, item := range items {
		var preceedingItem Item = nil
		if position > 0 {
			preceedingItem = items[position-1]
		}
		if !IsValidBreakpoint(preceedingItem, item) {
			continue
		}
		for activeNode := range activeNodes {
			// fmt.Println("here!")
			// fmt.Println("Considering break from ", activeNode.position, "to", position)
			// fmt.Println("Line width", lineData.GetWidth(activeNode.position, position))

			adjustmentRatio := calculateAdjustmentRatio(
				lineData.GetWidth(activeNode.position, position),
				lineData.GetShrinkability(activeNode.position, position),
				lineData.GetStrechability(activeNode.position, position),
				subsequentLineLengths,
			)

			// We can't break here, and we can't break in any future positions using this
			// node because the adjustmentRatio is only going to get worse
			if adjustmentRatio < -1 {
				println("Skipping, adjustment ratio too small", adjustmentRatio)
				delete(activeNodes, activeNode)
				continue
			}
			// We must add a break here, in which case previous active nodes are deleted
			if items[position].PenaltyCost() <= NegativeInfinity {
				delete(activeNodes, activeNode)
			}
			// This is the case when there is not enough material for the line.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				fmt.Println("Skipping", activeNode.position, "to", position, "(adjustment ratio too large)")
				continue
			}
			thisNode := node{
				position:     position,
				line:         activeNode.line + 1, // Todo: should be -1 in some cases
				fitnessClass: criteria.CalculateFitnessClass(adjustmentRatio),
			}
			newActiveNodes[thisNode] = true
			// NOTE: this is wrong! The item of interest if the one pointed to be active node
			preceedingItemIsFlaggedPenalty := false
			if preceedingItem != nil {
				preceedingItemIsFlaggedPenalty = false // preceedingItem.IsFlaggedPenalty()
			}
			demerits := criteria.CalculateDemerits(
				adjustmentRatio,
				thisNode.fitnessClass,
				activeNode.fitnessClass,
				item.PenaltyCost(),
				item.IsFlaggedPenalty(),
				preceedingItemIsFlaggedPenalty,
			) + nodeToMinDemerits[activeNode]
			println("Cost of going from", activeNode.position, "to", position, "is", demerits)
			minDemeritsSoFar := true
			if _, nodeAlreadyEncountered := nodeToPrevious[thisNode]; nodeAlreadyEncountered {
				minDemeritsSoFar = demerits < nodeToMinDemerits[thisNode]
			}
			//if nodeToPrevious[thisNode] != nil {
			//}
			if minDemeritsSoFar {
				print("Setting previous")
				nodeToPrevious[thisNode] = activeNode
				nodeToMinDemerits[thisNode] = demerits
			}
		}
		for newActiveNode := range newActiveNodes {
			// fmt.Println("Adding new active nodes")
			delete(newActiveNodes, newActiveNode)
			activeNodes[newActiveNode] = true
		}
	}
	if len(activeNodes) == 0 {
		return nil, &NoSolutionError{}
	}
	var bestNode node
	minDemerits := float64(-1)
	for activeNode := range activeNodes {
		if minDemerits < 0 || nodeToMinDemerits[activeNode] < minDemerits {
			minDemerits = nodeToMinDemerits[activeNode]
			bestNode = activeNode
		}
	}
	numBreakpoints := int64(0)
	for thisNode := bestNode; thisNode.position != -1; thisNode = nodeToPrevious[thisNode] {
		numBreakpoints++
	}
	breakpointIndices := make([]int, numBreakpoints)
	for thisNode := bestNode; thisNode.position != -1; thisNode = nodeToPrevious[thisNode] {
		breakpointIndices[numBreakpoints-1] = thisNode.position
		numBreakpoints--
	}
	return breakpointIndices, nil
}

// lineData is a data structure that facilitates computing the width, shrinkability
// and stretchability of a line (i.e., a contiguous subsequence of Items) in O(1).
// It incorporates the logic that a glue item at the end of a line does not contribute
// to any of these quantities, and that a penalty item only contributes if it is at the
// end.
type lineData struct {
	aggregateWidth          []int64
	aggregateShrinkability  []int64
	aggregateStretchibility []int64
	items                   []Item
}

func buildLineData(items []Item) *lineData {
	lineData := &lineData{
		aggregateWidth:          make([]int64, len(items)+1),
		aggregateShrinkability:  make([]int64, len(items)+1),
		aggregateStretchibility: make([]int64, len(items)+1),
		items:                   items,
	}
	lineData.aggregateWidth[0] = 0
	lineData.aggregateShrinkability[0] = 0
	lineData.aggregateStretchibility[0] = 0
	for position, item := range items {
		lineData.aggregateWidth[position+1] =
			lineData.aggregateWidth[position] +
				item.Width()
		lineData.aggregateShrinkability[position+1] =
			lineData.aggregateShrinkability[position] +
				item.Shrinkability()
		lineData.aggregateStretchibility[position+1] =
			lineData.aggregateStretchibility[position] +
				item.Stretchability()
	}
	return lineData
}

func (lineData *lineData) GetWidth(previousBreakpoint int, thisBreakPoint int) int64 {
	return lineData.aggregateWidth[thisBreakPoint+1] -
		lineData.aggregateWidth[previousBreakpoint+1] +
		lineData.items[thisBreakPoint].EndOfLineWidth() -
		lineData.items[thisBreakPoint].Width()
}

func (lineData *lineData) GetShrinkability(start int, end int) int64 {
	return lineData.aggregateShrinkability[end+1] -
		lineData.aggregateShrinkability[start+1] -
		lineData.items[end].Shrinkability()
}
func (lineData *lineData) GetStrechability(start int, end int) int64 {
	return lineData.aggregateStretchibility[end+1] -
		lineData.aggregateStretchibility[start+1] -
		lineData.items[end].Stretchability()
}

// NoSolutionError is returned if the problem has no solution satisfying the optimality contraints
type NoSolutionError struct{}

func (err *NoSolutionError) Error() string {
	return "There is no admissible solution given the provided optimiality criteria!"
}

func calculateAdjustmentRatio(
	lineWidth int64,
	lineShrinkability int64,
	lineStretchability int64,
	targetLineWidth int64,

) float64 {
	if lineWidth > targetLineWidth {
		// Division by 0, inf is allowed
		// fmt.Println("Why shrinking?", lineWidth, targetLineWidth, lineShrinkability)
		return float64(-lineWidth+targetLineWidth) / float64(lineShrinkability)
	}
	if lineWidth < targetLineWidth {
		// fmt.Println("Stretching", lineWidth, targetLineWidth, lineStretchability)
		return float64(-lineWidth+targetLineWidth) / float64(lineStretchability)
	}
	return 0
}

type node struct {
	position     int // NOTE: golang restrictions mean this will be the max length of items
	line         int64
	fitnessClass FitnessClass
	// Note: total width needs to account for penalty width of the breakpoint
}
