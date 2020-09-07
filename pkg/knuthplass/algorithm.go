package knuthplass

import (
	"fmt"
)

type LineLengths struct {
	initiaLengths     []int64
	subsequentLengths int64
}

func NewConstantLineLengths(constantLength int64) LineLengths {
	return LineLengths{initiaLengths: []int64{}, subsequentLengths: constantLength}
}

func (lineLengths *LineLengths) GetLength(lineIndex int) int64 {
	if lineIndex < len(lineLengths.initiaLengths) {
		return lineLengths.initiaLengths[lineIndex]
	}
	return lineLengths.subsequentLengths
}

func (lineLengths *LineLengths) GetNextIndex(lineIndex int, distinguishSubsequentLines bool) int {
	if distinguishSubsequentLines || lineIndex < len(lineLengths.initiaLengths) {
		return lineIndex + 1
	}
	return lineIndex
}

func KnuthPlassAlgorithm(
	lineData *ItemList,
	lineLengths LineLengths,
	criteria OptimalityCriteria,
) ([]int, error) {
	// Set of nodes
	activeNodes := make(map[node]bool)
	newActiveNodes := make(map[node]bool)
	firstNode := node{itemIndex: -1, lineIndex: 0, fitnessClass: 1}
	activeNodes[firstNode] = true

	// Data about the nodes. We don't include this data on the node object itself
	// as that would interfere with hashing
	nodeToPrevious := make(map[node]node)
	nodeToTotalDemerits := make(map[node]float64)

	for itemIndex := 0; itemIndex < lineData.Length(); itemIndex++ {
		precedingItem := lineData.Get(itemIndex - 1)
		item := lineData.Get(itemIndex)
		if !IsValidBreakpoint(precedingItem, item) {
			continue
		}
		for activeNode := range activeNodes {
			// fmt.Println("here!")
			// fmt.Println("Considering break from ", activeNode.itemIndex, "to", itemIndex)
			// fmt.Println("Line width", lineData.GetWidth(activeNode.itemIndex, itemIndex))
			thisLineIndex := lineLengths.GetNextIndex(activeNode.lineIndex, criteria.GetLooseness() != 0)
			thisLineItems := lineData.Slice(activeNode.itemIndex+1, itemIndex+1)
			adjustmentRatio := calculateAdjustmentRatio(
				thisLineItems.Width(),
				thisLineItems.Shrinkability(),
				thisLineItems.Stretchability(),
				lineLengths.GetLength(thisLineIndex),
			)
			println("Adjustment ratio", activeNode.itemIndex, "to", itemIndex, adjustmentRatio)

			// We can't break here, and we can't break in any future positions using this
			// node because the adjustmentRatio is only going to get worse
			if adjustmentRatio < -1 {
				println("Skipping, adjustment ratio too small", adjustmentRatio)
				delete(activeNodes, activeNode)
				continue
			}
			// We must add a break here, in which case previous active nodes are deleted
			if item.PenaltyCost() <= NegativeInfinity {
				delete(activeNodes, activeNode)
			}
			// This is the case when there is not enough material for the lineIndex.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				fmt.Println("Skipping", activeNode.itemIndex, "to", itemIndex, "(adjustment ratio too large)")
				continue
			}
			thisNode := node{
				itemIndex:    itemIndex,
				lineIndex:    thisLineIndex,
				fitnessClass: criteria.CalculateFitnessClass(adjustmentRatio),
			}
			newActiveNodes[thisNode] = true
			// NOTE: this is wrong! The item of interest if the one pointed to be active node
			// TODO: fix this with a test
			preceedingItemIsFlaggedPenalty := false
			if precedingItem != nil {
				preceedingItemIsFlaggedPenalty = false // precedingItem.IsFlaggedPenalty()
			}
			demerits := criteria.CalculateDemerits(
				adjustmentRatio,
				thisNode.fitnessClass,
				activeNode.fitnessClass,
				item.PenaltyCost(),
				item.IsFlaggedPenalty(),
				preceedingItemIsFlaggedPenalty,
			) + nodeToTotalDemerits[activeNode]
			println("Cost of going from", activeNode.itemIndex, "to", itemIndex, "is", demerits)
			minDemeritsSoFar := true
			if _, nodeAlreadyEncountered := nodeToPrevious[thisNode]; nodeAlreadyEncountered {
				minDemeritsSoFar = demerits < nodeToTotalDemerits[thisNode]
			}
			//if nodeToPrevious[thisNode] != nil {
			//}
			if minDemeritsSoFar {
				nodeToPrevious[thisNode] = activeNode
				nodeToTotalDemerits[thisNode] = demerits
			}
		}
		for newActiveNode := range newActiveNodes {
			// fmt.Println("Adding new active nodes")
			delete(newActiveNodes, newActiveNode)
			activeNodes[newActiveNode] = true
		}
	}
	if len(activeNodes) == 0 {
		// TODO, give something back in this case
		return nil, &NoSolutionError{}
	}
	var bestNode node
	minDemerits := float64(-1)
	for activeNode := range activeNodes {
		if minDemerits < 0 || nodeToTotalDemerits[activeNode] < minDemerits {
			minDemerits = nodeToTotalDemerits[activeNode]
			bestNode = activeNode
		}
	}
	numBreakpoints := int64(0)
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = nodeToPrevious[thisNode] {
		numBreakpoints++
	}
	breakpointIndices := make([]int, numBreakpoints)
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = nodeToPrevious[thisNode] {
		breakpointIndices[numBreakpoints-1] = thisNode.itemIndex
		numBreakpoints--
	}
	return breakpointIndices, nil
}

// NoSolutionError is returned if the problem has no solution satisfying the optimality contraints
type NoSolutionError struct {
	// TODO: add data on which lines failed
}

func (err *NoSolutionError) Error() string {
	return "There is no admissible solution given the provided optimiality criteria!"
}

func calculateAdjustmentRatio(
	lineWidth int64,
	lineShrinkability int64,
	lineStretchability int64,
	targetLineWidth int64,

) float64 {
	switch {
	case lineWidth > targetLineWidth:
		// Division by 0, inf is allowed
		// fmt.Println("Why shrinking?", lineWidth, targetLineWidth, lineShrinkability)
		return float64(-lineWidth+targetLineWidth) / float64(lineShrinkability)
	case lineWidth < targetLineWidth:
		if lineStretchability == InfiniteStretchability {
			return 0
		}
		// fmt.Println("Stretching", lineWidth, targetLineWidth, lineStretchability)
		return float64(-lineWidth+targetLineWidth) / float64(lineStretchability)
	default:
		return 0
	}
}

type node struct {
	itemIndex    int
	lineIndex    int
	fitnessClass FitnessClass
}
