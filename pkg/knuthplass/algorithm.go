// Package knuthplass contains an implementation of the Knuth-Plass line breaking algorithm, as well as definitions
// for the data structures needed to use it.
//
// Users of this package must first create:
//
// - a slice of Item types representing the problem at hand, and then use these to initialize an ItemList.
//
// - a LineLengths type that describes the line lengths in the paragraph.
//
// - an OptimalityCriteria type that describes the optimization criteria. The type TexOptimalityCriteria provides
// the optimality criteria as used in the Tex typesetting system.
//
// With these, the Knuth-Plass algorithm can be run using the CalculateBreakpoints function.
// After this, the lines can be set (i.e., the exact width of each input Item can be fixed) using the TBD function.
package knuthplass

import "fmt"

// LineLengths provides the lengths of the lines in a paragraph.
//
// A bad implementation of LineLengths may result in the Knuth-Plass algorithm running an order of magnitude
// slower, as explained below. Therefore, instead of writing a new type that implements this interface, prefer to use
// the functions NewConstantLineLengths and NewVariableLineLengths if they suit the given use case.
//
// Otherwise, read the following optimization discussion carefully.
// In general, the Knuth-Plass algorithm runs in O(L*N*M) where L is the number of lines the paragraph, N is the number
// of items in the input, and M is approximately the number of items that can fit on each line.
// There exists a simple optimization in the original paper of Knuth and Plass such that if the line lengths
// are all constant, the running time is O(M*N); i.e., effectively L=1.
// In this package's implementation of Knuth-Plass, this optimization is achieved by using a smart LineLengths
// implementation.
//
// The following observation leads to the optimization.
// Let A be the index of a line in the paragraph, and let B be the index of the following line.
// In general this means B=A+1.
// However, the Knuth-Plass algorithm doesn't require this.
// The only constraint the algorithm imposes is that the list of line lengths starting after index A
// is equal to the list of line lengths starting at index B.
// If B=A+1 this is trivially true.
// But consider the case when all the line lengths are the same.
// In this case, regardless of the starting point, the list of line lengths is just the same
// constant list.
// We can thus set index B to be *equal* to index A, because the the two lists in the constraint will be the
// same constant list and in particular the same.
// In fact, for a paragraph of constant lengths all lines have the same "index".
//
// The LineLengths interface refers to the "index" as the pseudo index to indicate that it is not, in general,
// the actual index of a line in a paragraph but rather the number satisfying the constraint above.
// Using the LineLengths interface, the running time of the algorithm is O(U*N*M) where U is the number of
// unique line pseudo indexes in the paragraph.
type LineLengths interface {
	// GetLength returns the length of the line at the provided pseudo index.
	GetLength(pseudoIndex int) int64

	// GetNextPseudoIndex returns the pseudo index of the line after the line specified by the given pseudo index.
	GetNextPseudoIndex(pseudoIndex int) int
}

// NewConstantLineLengths constructs a new LineLengths data structure where each line has the same length.
func NewConstantLineLengths(constantLength int64) LineLengths {
	return &basicLineLengthsImpl{initialLengths: []int64{}, subsequentLengths: constantLength}
}

// NewVariableLineLengths constructs a new LineLengths data structure where the first lines have the variable
// lengths given in the lengths parameter, and all subsequent lines have the length given in subsequentLengths.
func NewVariableLineLengths(lengths []int64, subsequentLengths int64) LineLengths {
	return &basicLineLengthsImpl{initialLengths: lengths, subsequentLengths: subsequentLengths}
}

type basicLineLengthsImpl struct {
	initialLengths    []int64
	subsequentLengths int64
}

func (lineLengths *basicLineLengthsImpl) GetLength(pseudoIndex int) int64 {
	if pseudoIndex < len(lineLengths.initialLengths) {
		return lineLengths.initialLengths[pseudoIndex]
	}
	return lineLengths.subsequentLengths
}

func (lineLengths *basicLineLengthsImpl) GetNextPseudoIndex(pseudoIndex int) int {
	if pseudoIndex < len(lineLengths.initialLengths) {
		return pseudoIndex + 1
	}
	return pseudoIndex
}

// CalculateBreakpointsResult stores the result of the CalculateBreakpoints algorithm.
type CalculateBreakpointsResult struct {
	// Breakpoints contains the indices of the items in the ItemList where the optimal line breaks occur.
	//
	// Note that in the case when Err is not nil and the fallbackToIllegalSolution parameter to CalculateBreakpoints was
	// false, Breakpoints will be nil.
	Breakpoints []int

	// Err contains an error if the Knuth-Plass problem could not be solved subject to the provided constraints.
	Err *NoSolutionError

	// Logger contains a BreakpointLogger with results of the internal calculations performed during the function.
	// If the enableLogging parameter to CalculateBreakpoints was false, the logger will be nil.
	Logger *BreakpointLogger
}

// CalculateBreakpoints is the package's implementation of the Knuth-Plass algorithm.
// It finds the optimal sequence of Breakpoints in an ItemList
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
	firstNode := activeNodes.add(nodeID{itemIndex: -1, lineIndex: -1, fitnessClass: 0})
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
			// We refer to the target node using a nodeID object. This is because the target node is not in the
			// active set, and hence there is no node object corresponding it.
			targetNode      nodeID
			adjustmentRatio float64
		}
		var edges []edge
		var fallbackEdges []edge
		var sourceNodesToDeactivate []*node

		allActiveNodesToBeDeactivated := true
		for _, sourceNode := range activeNodes.idToNode {
			var thisLineIndex int
			if criteria.GetLooseness() != 0 {
				thisLineIndex = sourceNode.lineIndex + 1
			} else {
				thisLineIndex = lineLengths.GetNextPseudoIndex(sourceNode.lineIndex)
			}
			thisLineItems := itemList.Slice(sourceNode.itemIndex+1, itemIndex+1)
			adjustmentRatio := calculateAdjustmentRatio(
				thisLineItems.Width(),
				thisLineItems.Shrinkability(),
				thisLineItems.Stretchability(),
				lineLengths.GetLength(thisLineIndex),
			)
			targetNode := nodeID{
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
				fmt.Println("falling back!")
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
			if logger != nil {
				logger.LineDemeritsTable.AddCell(edge.sourceNode, edge.targetNode, lineDemerits)
				logger.TotalDemeritsTable.AddCell(
					edge.sourceNode,
					edge.targetNode,
					lineDemerits+edge.sourceNode.demerits,
				)
			}
			targetNode := activeNodes.add(edge.targetNode)
			updateTargetNodeIfNewSourceNodeIsBetter(targetNode, edge.sourceNode, lineDemerits)
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
		updateTargetNodeIfNewSourceNodeIsBetter(finalPseudoNode, activeNode, 0)
	}
	var err *NoSolutionError
	if len(problematicItemIndices) != 0 {
		err = &NoSolutionError{ProblematicItemIndices: problematicItemIndices}
	}
	return CalculateBreakpointsResult{
		Breakpoints: buildBreakpoints(finalPseudoNode.prevNode),
		Err:         err,
		Logger:      logger,
	}
}

// buildBreakpoints returns a slice of breakpoint indices, where the provided node is the final breakpoint.
func buildBreakpoints(node *node) []int {
	if node == nil {
		return nil
	}
	numBreakpoints := 0
	for thisNode := node; thisNode.prevNode != nil; thisNode = thisNode.prevNode {
		numBreakpoints++
	}
	breakpoints := make([]int, numBreakpoints)
	for thisNode := node; thisNode.prevNode != nil; thisNode = thisNode.prevNode {
		breakpoints[numBreakpoints-1] = thisNode.itemIndex
		numBreakpoints--
	}
	return breakpoints
}

// updateTargetNodeIfNewSourceNodeIsBetter changes the previous node of the targetNode to be the provided
// candidatePrevNode if that node results in the targetNode having fewer demerits.
func updateTargetNodeIfNewSourceNodeIsBetter(targetNode *node, candidatePrevNode *node, edgeDemerits float64) {
	totalDemerits := candidatePrevNode.demerits + edgeDemerits
	switch true {
	case targetNode.prevNode == nil || totalDemerits < targetNode.demerits:
		targetNode.prevNode = candidatePrevNode
		targetNode.demerits = totalDemerits
	case totalDemerits == targetNode.demerits:
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

type node struct {
	itemIndex    int
	lineIndex    int
	fitnessClass FitnessClass
	demerits     float64
	prevNode     *node
}

type nodeID struct {
	itemIndex    int
	lineIndex    int
	fitnessClass FitnessClass
}

func (node *node) id() nodeID {
	return nodeID{
		itemIndex:    node.itemIndex,
		lineIndex:    node.lineIndex,
		fitnessClass: node.fitnessClass,
	}
}

type nodeSet struct {
	idToNode map[nodeID]*node
}

func newNodeSet() nodeSet {
	return nodeSet{idToNode: map[nodeID]*node{}}
}

func (set *nodeSet) add(nodeID nodeID) *node {
	_, exists := set.idToNode[nodeID]
	if !exists {
		set.idToNode[nodeID] = &node{
			itemIndex:    nodeID.itemIndex,
			lineIndex:    nodeID.lineIndex,
			fitnessClass: nodeID.fitnessClass,
		}
	}
	return set.idToNode[nodeID]
}

func (set *nodeSet) delete(node *node) {
	delete(set.idToNode, node.id())
}

func (set *nodeSet) len() int {
	return len(set.idToNode)
}
