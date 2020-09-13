// Package knuthplass contains an implementation of the Knuth-Plass line breaking algorithm, as well as definitions
// for the data structures needed to use it.
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

// CalculateBreakpointsResult stores the result of the CalculateBreakpoints algorithm.
type CalculateBreakpointsResult struct {
	// Breakpoints contains the indices of the items in the ItemList where the optimal line breaks occur.
	//
	// Note that in the case when Err is not nil and the returnBestEffort parameter to CalculateBreakpoints was
	// false, Breakpoints will be nil.
	Breakpoints []int

	// Err contains an error if the Knuth-Plass problem could not be solved subject to the provided constraints.
	Err error

	// Logger contains a BreakpointLogger with results of the internal calculations performed during the function.
	// If the enableLogging parameter to CalculateBreakpoints was false, the logger will be nil.
	Logger *BreakpointLogger
}

// CalculateBreakpoints is an implementation of the Knuth-Plass algorithm. It uses the algorithm to find the optimal
// sequence of Breakpoints in an ItemList
// such that when the resulting lines are set using the provided LineLengths the total number of demerits as calculated
// using the OptimalityCriteria is minimized.
//
// The Knuth-Plass problem may not have a solution. This occurs when there is no way to set the paragraph such
// that each line's adjustment ratio is in the range (-1, OptimalityCriteria.GetMaxAdjustmentRatio()). In this
// case CalculateBreakpoints can do two things. If the parameter returnBestEffort is false, the function terminates with
// an error as soon as the non-existence of a solution is discovered. If returnBestEffort is true, the function will
// continue and return a full set of breakpoints; however, one or more of the lines will necessarily be under-full
// (adjustment ratio less than -1) or over-full (adjustment ratio bigger than the max). An error in this case will be
// returned, and will describe which lines are affected and whether they're under- or over-full.
//
// Consult the documentation on the output type CalculateBreakpointsResult for more information on this function
// and its parameters.
func CalculateBreakpoints(
	itemList *ItemList,
	lineLengths LineLengths,
	criteria OptimalityCriteria,
	enableLogging bool,
) CalculateBreakpointsResult {
	// Active nodes is used as a set of nodes; the value is ignored.
	// activeNodes := make(map[node]bool)
	activeNodes := &activeNodeSet{}
	firstNode := node{itemIndex: -1, lineIndex: -1, fitnessClass: 0}
	activeNodes.initialize(firstNode)
	// activeNodes[firstNode] = true

	// Data about the nodes. We don't include this data on the node object itself as that would interfere with hashing.
	nodeToPrevious := make(map[node]*node)
	nodeToTotalDemerits := make(map[node]float64)

	var logger *BreakpointLogger
	if enableLogging {
		logger = NewBreakpointLogger()
	} else {
		logger = nil
	}

	for itemIndex := 0; itemIndex < itemList.Length(); itemIndex++ {
		item := itemList.Get(itemIndex)
		if !item.IsValidBreakpoint(itemList.Get(itemIndex - 1)) {
			continue
		}
		for activeNode := range activeNodes.nodesInSet {
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
				// TODO: set the adjustmentRatio to -1 and continue with this overfull box
				// We set it to -1 as that will the way it is set

				activeNodes.markForDeletion(activeNode)
				continue
			}
			// We must add a break here, in which case previous active nodes are deleted
			if item.BreakpointPenalty() <= NegInfBreakpointPenalty {
				activeNodes.markForDeletion(activeNode)
			}
			// This is the case when there is not enough material for the line.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				// TODO: under-full boxes?
				continue
			}

			totalDemerits := criteria.CalculateDemerits(
				adjustmentRatio,
				thisNode.fitnessClass,
				activeNode.fitnessClass,
				item.BreakpointPenalty(),
				item.IsFlaggedBreakpoint(),
				IsNullableItemFlaggedBreakpoint(itemList.Get(activeNode.itemIndex)),
			) + nodeToTotalDemerits[activeNode]
			if logger != nil {
				logger.LineDemeritsTable.AddCell(activeNode, thisNode, totalDemerits-nodeToTotalDemerits[activeNode])
				logger.TotalDemeritsTable.AddCell(activeNode, thisNode, totalDemerits)
			}

			activeNodes.markForAddition(thisNode)
			nodeToPrevious[thisNode], nodeToTotalDemerits[thisNode] = selectBestNode(
				nodeToPrevious[thisNode],
				nodeToTotalDemerits[thisNode],
				activeNode,
				totalDemerits,
			)
		}
		activeNodes.commitMarkedChanges()
	}
	return buildResult(
		activeNodes.nodesInSet,
		nodeToPrevious,
		nodeToTotalDemerits,
		logger,
	)
}

func buildResult(
	activeNodes map[node]bool,
	nodeToPrevious map[node]*node,
	nodeToTotalDemerits map[node]float64,
	logger *BreakpointLogger,
) CalculateBreakpointsResult {
	if len(activeNodes) == 0 {
		// TODO, give something back in this case
		return CalculateBreakpointsResult{nil, &NoSolutionError{}, logger}
	}
	var bestNode *node
	var minDemerits float64
	for activeNode := range activeNodes {
		bestNode, minDemerits = selectBestNode(
			bestNode,
			minDemerits,
			activeNode,
			nodeToTotalDemerits[activeNode],
		)
	}
	if bestNode == nil {
		panic(0)
	}
	numBreakpoints := 0
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = nodeToPrevious[*thisNode] {
		numBreakpoints++
	}
	breakpointIndices := make([]int, numBreakpoints)
	for thisNode := bestNode; thisNode.itemIndex != -1; thisNode = nodeToPrevious[*thisNode] {
		breakpointIndices[numBreakpoints-1] = thisNode.itemIndex
		numBreakpoints--
	}
	return CalculateBreakpointsResult{breakpointIndices, nil, logger}
}

func selectBestNode(node1 *node, demerits1 float64, node2 node, demerits2 float64) (*node, float64) {
	switch true {
	case node1 == nil:
		return &node2, demerits2
	case demerits1 > demerits2:
		return &node2, demerits2
	case demerits1 < demerits2:
		return node1, demerits1
	}
	// This is the case when the demerits are the same. In order to have a deterministic algorithm, we select
	// the "smallest" node as determined by the following function.
	return selectSmallerNode(node1, &node2), demerits1
}

// selectSmallerNode returns the "smaller" of two nodes, where the composite ordering considers the fields itemIndex,
// fitnessClass and lineIndex in tha order.
func selectSmallerNode(node1 *node, node2 *node) *node {
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
	// Which breakpoint and whether underfful or overfull
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
		return float64(-lineWidth+targetLineWidth) / float64(lineShrinkability)
	case lineWidth < targetLineWidth:
		if lineStretchability == InfiniteStretchability {
			return 0
		}
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

type activeNodeSet struct {
	nodesInSet map[node]bool
	nodesToDelete map[node]bool
	nodesToAdd map[node]bool
}

func (set *activeNodeSet) initialize(startingNode node) {
	set.nodesInSet = make(map[node]bool)
	set.nodesToDelete = make(map[node]bool)
	set.nodesToAdd = make(map[node]bool)
	set.nodesInSet[startingNode] = true
}

func (set *activeNodeSet) markForDeletion(node node) {
	set.nodesToDelete[node] = true
}

func (set *activeNodeSet) markForAddition(node node) {
	set.nodesToAdd[node] = true
}

func (set *activeNodeSet) commitMarkedChanges() {
	for nodeToDelete := range set.nodesToDelete {
		delete(set.nodesInSet, nodeToDelete)
	}
	for nodeToAdd := range set.nodesToAdd {
		set.nodesInSet[nodeToAdd] = true
	}
	set.dropMarkedChanges()
}

func (set *activeNodeSet) dropMarkedChanges() {
	set.nodesToDelete = make(map[node]bool)
	set.nodesToAdd = make(map[node]bool)
}