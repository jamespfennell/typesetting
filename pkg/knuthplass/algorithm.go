// Package knuthplass contains an implementation of the Knuth-Plass line breaking algorithm, as well as definitions
// for the data structures needed to use it.
package knuthplass

import "fmt"

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
	Err *NoSolutionError

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
// an error as soon as the non-existence of a solution is discovered. If fallbackToIllegalSolution is true, the function
// will continue and return a full set of breakpoints; however, one or more of the lines will necessarily be under-full
// (adjustment ratio less than -1) or over-full (adjustment ratio bigger than the max). An error in this case will be
// returned, and will describe which lines are affected and whether they're under- or over-full.
//
// Consult the documentation on the output type CalculateBreakpointsResult for more information on this function
// and its parameters.
func CalculateBreakpoints(
	itemList *ItemList,
	lineLengths LineLengths,
	criteria OptimalityCriteria,
	fallbackToIllegalSolution bool,
	enableLogging bool,
) CalculateBreakpointsResult {
	activeNodes := newNodeSet()
	firstNode := activeNodes.add(nodeId{itemIndex: -1, lineIndex: -1, fitnessClass: 0})
	firstNode.demerits = 0
	var problematicItemIndices []int

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
		type edge struct {
			sourceNode *node
			// We refer to the target node using a nodeId object. This is because the target node is not in the
			// active set, and hence there is no node object corresponding it.
			targetNode      nodeId
			adjustmentRatio float64
		}
		var edges []edge
		var fallbackEdges []edge
		var sourceNodesToDeactivate []*node

		allActiveNodesToBeDeactivated := true
		for _, sourceNode := range activeNodes.idToNode {
			thisLineIndex := lineLengths.GetNextIndex(sourceNode.lineIndex, criteria.GetLooseness() != 0)
			thisLineItems := itemList.Slice(sourceNode.itemIndex+1, itemIndex+1)
			adjustmentRatio := calculateAdjustmentRatio(
				thisLineItems.Width(),
				thisLineItems.Shrinkability(),
				thisLineItems.Stretchability(),
				lineLengths.GetLength(thisLineIndex),
			)
			targetNode := nodeId{
				itemIndex:    itemIndex,
				lineIndex:    thisLineIndex,
				fitnessClass: criteria.CalculateFitnessClass(adjustmentRatio),
			}
			if logger != nil {
				logger.AdjustmentRatiosTable.AddCell(sourceNode, targetNode, adjustmentRatio)
			}
			if adjustmentRatio < -1 || item.BreakpointPenalty() <= NegInfBreakpointPenalty {
				sourceNodesToDeactivate = append(sourceNodesToDeactivate, sourceNode)
			} else {
				allActiveNodesToBeDeactivated = false
			}
			thisEdge := edge{sourceNode: sourceNode, targetNode: targetNode, adjustmentRatio: adjustmentRatio}
			switch true {
			case adjustmentRatio >= -1 && adjustmentRatio <= criteria.GetMaxAdjustmentRatio():
				edges = append(edges, thisEdge)
			case len(edges) == 0 && allActiveNodesToBeDeactivated:
				fallbackEdges = append(fallbackEdges, thisEdge)
			}
		}
		// If there are no legal edges and after this iteration there will be no active nodes, there is no legal
		// solution to this Knuth-Plass problem.
		if len(edges) == 0 && allActiveNodesToBeDeactivated {
			problematicItemIndices = append(problematicItemIndices, itemIndex)
			if fallbackToIllegalSolution {
				edges = fallbackEdges
			}
		}
		for _, edge := range edges {
			lineDemerits := criteria.CalculateDemerits(
				edge.adjustmentRatio,
				edge.targetNode.fitnessClass,
				edge.sourceNode.fitnessClass,
				item.BreakpointPenalty(),
				item.IsFlaggedBreakpoint(),
				IsNullableItemFlaggedBreakpoint(itemList.Get(edge.sourceNode.itemIndex)),
			)
			totalDemerits := lineDemerits + edge.sourceNode.demerits
			if logger != nil {
				logger.LineDemeritsTable.AddCell(edge.sourceNode, edge.targetNode, lineDemerits)
				logger.TotalDemeritsTable.AddCell(edge.sourceNode, edge.targetNode, totalDemerits)
			}
			targetNode := activeNodes.add(edge.targetNode)
			updateTargetNodeIfNewSourceNodeIsBetter(targetNode, edge.sourceNode, totalDemerits)
		}
		for _, activeNode := range sourceNodesToDeactivate {
			activeNodes.delete(activeNode)
		}
		if activeNodes.len() == 0 {
			break
		}
	}
	finalPseudoNode := &node{}
	for _, activeNode := range activeNodes.idToNode {
		updateTargetNodeIfNewSourceNodeIsBetter(finalPseudoNode, activeNode, activeNode.demerits)
	}
	numBreakpoints := 0
	for thisNode := finalPseudoNode.prevNode; thisNode.prevNode != nil; thisNode = thisNode.prevNode {
		numBreakpoints++
	}
	result := CalculateBreakpointsResult{
		Breakpoints: make([]int, numBreakpoints),
		Err:         nil,
		Logger:      logger,
	}
	for thisNode := finalPseudoNode.prevNode; thisNode.prevNode != nil; thisNode = thisNode.prevNode {
		result.Breakpoints[numBreakpoints-1] = thisNode.itemIndex
		numBreakpoints--
	}
	// TODO: test this case
	if len(problematicItemIndices) != 0 {
		result.Err = &NoSolutionError{ProblematicItemIndices: problematicItemIndices}
	}
	return result
}

// updateTargetNodeIfNewSourceNodeIsBetter changes the previous node of the targetNode to be the provided
// candidatePrevNode if that node results in the targetNode having fewer demerits.
func updateTargetNodeIfNewSourceNodeIsBetter(targetNode *node, candidatePrevNode *node, candidateDemerits float64) {
	switch true {
	case targetNode.prevNode == nil || candidateDemerits < targetNode.demerits:
		targetNode.prevNode = candidatePrevNode
		targetNode.demerits = candidateDemerits
	case candidateDemerits == targetNode.demerits:
		targetNode.prevNode = selectSmallerNode(targetNode.prevNode, candidatePrevNode)
	}
}

// selectSmallerNode returns the "smaller" of two nodes, where the composite ordering considers the fields itemIndex,
// fitnessClass and lineIndex in that order. This ordering is used to make the algorithm deterministic; otherwise,
// the result may depend on the non-deterministic iteration order of active nodes.
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
	ProblematicItemIndices []int
}

func (err *NoSolutionError) Error() string {
	return fmt.Sprintf(
		"There is no admissible solution given the provided optimality criteria. "+
			"Non-legal breakpoints occur at the following item indices: %v",
		err.ProblematicItemIndices,
	)

}

func calculateAdjustmentRatio(
	lineWidth int64,
	lineShrinkability int64,
	lineStretchability int64,
	targetLineWidth int64,
) float64 {
	switch {
	case lineWidth > targetLineWidth:
		return float64(-lineWidth+targetLineWidth) / float64(lineShrinkability)
	case lineWidth < targetLineWidth:
		if lineStretchability >= InfiniteStretchability {
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
	demerits     float64
	prevNode     *node
}

type nodeId struct {
	itemIndex    int
	lineIndex    int
	fitnessClass FitnessClass
}

func (node *node) id() nodeId {
	return nodeId{
		itemIndex:    node.itemIndex,
		lineIndex:    node.lineIndex,
		fitnessClass: node.fitnessClass,
	}
}

type nodeSet struct {
	idToNode map[nodeId]*node
}

func newNodeSet() nodeSet {
	return nodeSet{idToNode: map[nodeId]*node{}}
}

func (set *nodeSet) add(nodeId nodeId) *node {
	_, exists := set.idToNode[nodeId]
	if !exists {
		set.idToNode[nodeId] = &node{
			itemIndex:    nodeId.itemIndex,
			lineIndex:    nodeId.lineIndex,
			fitnessClass: nodeId.fitnessClass,
		}
	}
	return set.idToNode[nodeId]
}

func (set *nodeSet) delete(node *node) {
	delete(set.idToNode, node.id())
}

func (set *nodeSet) len() int {
	return len(set.idToNode)
}
