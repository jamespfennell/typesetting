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

	// lineData := buildLineData(items)

	// Set of nodes
	activeNodes := make(map[node]bool)
	newActiveNodes := make(map[node]bool)
	firstNode := node{position: -1, line: 0, fitnessClass: 1}
	activeNodes[firstNode] = true

	// Data about the nodes. We don't include this data on the node object itself
	// as that would interfere with hashing
	nodeToPrevious := make(map[node]node)
	// Rename nodeToTotalDemerits
	nodeToMinDemerits := make(map[node]float64)

	for position := 0; position < lineData.Length(); position++ {
		preceedingItem := lineData.Get(position - 1)
		item := lineData.Get(position)
		if !IsValidBreakpoint(preceedingItem, item) {
			continue
		}
		for activeNode := range activeNodes {
			// fmt.Println("here!")
			// fmt.Println("Considering break from ", activeNode.position, "to", position)
			// fmt.Println("Line width", lineData.GetWidth(activeNode.position, position))
			thisLineIndex := lineLengths.GetNextIndex(activeNode.line, criteria.GetLooseness() != 0)
			thisLineItems := lineData.Slice(activeNode.position+1, position+1)
			adjustmentRatio := calculateAdjustmentRatio(
				thisLineItems.Width(),
				thisLineItems.Shrinkability(),
				thisLineItems.Stretchability(),
				lineLengths.GetLength(thisLineIndex),
			)
			println("Adjustment ratio", activeNode.position, "to", position, adjustmentRatio)

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
			// This is the case when there is not enough material for the line.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				fmt.Println("Skipping", activeNode.position, "to", position, "(adjustment ratio too large)")
				continue
			}
			thisNode := node{
				position:     position,
				line:         thisLineIndex,
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
		// TODO, give something back in this case
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
	position     int // NOTE: golang restrictions mean this will be the max length of items
	line         int
	fitnessClass FitnessClass
	// Note: total width needs to account for penalty width of the breakpoint
}
