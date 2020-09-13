package knuthplass

type LineLengths struct {
	initialLengths    []int64
	subsequentLengths int64
}

func NewConstantLineLengths(constantLength int64) LineLengths {
	return LineLengths{initialLengths: []int64{}, subsequentLengths: constantLength}
}

func (lineLengths *LineLengths) GetLength(lineIndex int) int64 {
	if lineIndex < len(lineLengths.initialLengths) {
		return lineLengths.initialLengths[lineIndex]
	}
	return lineLengths.subsequentLengths
}

func (lineLengths *LineLengths) GetNextIndex(lineIndex int, distinguishSubsequentLines bool) int {
	if distinguishSubsequentLines || lineIndex < len(lineLengths.initialLengths) {
		return lineIndex + 1
	}
	return lineIndex
}

type CalculateBreakpointsResult struct {
	breakpoints []int
	Err         error
	logger      *BreakpointLogger
}

func CalculateBreakpoints(
	itemList *ItemList,
	lineLengths LineLengths,
	criteria OptimalityCriteria,
) CalculateBreakpointsResult {
	// Set of nodes
	activeNodes := make(map[node]bool)
	newActiveNodes := make(map[node]bool)
	firstNode := node{itemIndex: -1, lineIndex: 0, fitnessClass: 1}
	activeNodes[firstNode] = true

	// Data about the nodes. We don't include this data on the node object itself
	// as that would interfere with hashing
	nodeToPrevious := make(map[node]*node)
	nodeToTotalDemerits := make(map[node]float64)

	var loggingEnabled bool = true
	var logger *BreakpointLogger
	if loggingEnabled {
		logger = NewBreakpointLogger()
	} else {
		logger = nil
	}

	for itemIndex := 0; itemIndex < itemList.Length(); itemIndex++ {
		precedingItem := itemList.Get(itemIndex - 1)
		item := itemList.Get(itemIndex)
		if !item.IsValidBreakpoint(precedingItem) {
			continue
		}
		for activeNode := range activeNodes {
			thisLineIndex := lineLengths.GetNextIndex(activeNode.lineIndex, criteria.GetLooseness() != 0)
			thisLineItems := itemList.Slice(activeNode.itemIndex+1, itemIndex+1)
			adjustmentRatio := calculateAdjustmentRatio(
				thisLineItems.Width(),
				thisLineItems.Shrinkability(),
				thisLineItems.Stretchability(),
				lineLengths.GetLength(thisLineIndex),
			)
			thisNode := node{
				itemIndex:    itemIndex,
				lineIndex:    thisLineIndex,
				fitnessClass: criteria.CalculateFitnessClass(adjustmentRatio),
			}
			if logger != nil {
				logger.AdjustmentRatiosTable.AddCell(activeNode, thisNode, adjustmentRatio)
			}

			// We can't break here, and we can't break in any future positions using this
			// node because the adjustmentRatio is only going to get worse
			if adjustmentRatio < -1 {
				delete(activeNodes, activeNode)
				continue
			}
			// We must add a break here, in which case previous active nodes are deleted
			if item.BreakpointPenalty() <= NegInfBreakpointPenalty {
				delete(activeNodes, activeNode)
			}
			// This is the case when there is not enough material for the line.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				continue
			}

			newActiveNodes[thisNode] = true

			var precedingItemIsFlaggedPenalty bool
			if activeNode.itemIndex != -1 {
				precedingItemIsFlaggedPenalty = itemList.Get(activeNode.itemIndex).IsFlaggedBreakpoint()
			} else {
				precedingItemIsFlaggedPenalty = false
			}

			totalDemerits := criteria.CalculateDemerits(
				adjustmentRatio,
				thisNode.fitnessClass,
				activeNode.fitnessClass,
				item.BreakpointPenalty(),
				item.IsFlaggedBreakpoint(),
				precedingItemIsFlaggedPenalty,
			) + nodeToTotalDemerits[activeNode]
			if logger != nil {
				logger.LineDemeritsTable.AddCell(activeNode, thisNode, totalDemerits-nodeToTotalDemerits[activeNode])
				logger.TotalDemeritsTable.AddCell(activeNode, thisNode, totalDemerits)
			}

			nodeToPrevious[thisNode], nodeToTotalDemerits[thisNode] = betterOption(
				nodeToPrevious[thisNode],
				nodeToTotalDemerits[thisNode],
				activeNode,
				totalDemerits,
			)
		}
		for newActiveNode := range newActiveNodes {
			delete(newActiveNodes, newActiveNode)
			activeNodes[newActiveNode] = true
		}
	}
	if len(activeNodes) == 0 {
		// TODO, give something back in this case
		return CalculateBreakpointsResult{nil, &NoSolutionError{}, logger}
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
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = *nodeToPrevious[thisNode] {
		numBreakpoints++
	}
	breakpointIndices := make([]int, numBreakpoints)
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = *nodeToPrevious[thisNode] {
		breakpointIndices[numBreakpoints-1] = thisNode.itemIndex
		numBreakpoints--
	}
	return CalculateBreakpointsResult{breakpointIndices, nil, logger}
}

func betterOption(node1 *node, demerits1 float64, node2 node, demerits2 float64) (*node, float64) {
	switch true {
	case node1 == nil:
		return &node2, demerits2
	case demerits1 > demerits2:
		return &node2, demerits2
	case demerits1 < demerits2:
		return node1, demerits1
	}
	// This is the case when the demerits are the same
	return smallerNode(node1, &node2), demerits1
}

func smallerNode(node1 *node, node2 *node) *node {
	switch true {
	case (*node1).itemIndex < (*node2).itemIndex:
		return node1
	case (*node1).itemIndex > (*node2).itemIndex:
		return node2
	case (*node1).fitnessClass < (*node2).fitnessClass:
		return node1
	case (*node1).fitnessClass > (*node2).fitnessClass:
		return node2
	case (*node1).lineIndex < (*node2).lineIndex:
		return node2
	}
	return node1
}

// NoSolutionError is returned if the problem has no solution satisfying the optimality constraints
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
