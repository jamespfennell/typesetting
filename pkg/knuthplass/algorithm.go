package knuthplass

func KnuthPlassAlgorithm(
	items []Item,
	startingLineLengths []int64,
	subsequentLineLengths int64,
	criteria OptimalityCriteria,
) ([]int64, error) {

	// During the algorithm we need to know the width, shrinkability and stretchability
	// of lines; i.e., contiguous subsequences of items. The following slices allow us to
	// compute these quantities in O(n) time: the width of a line starting wtih item k
	// and ending with item l
	// is widthSoFar[l+1] - widthSoFar[k].
	// TODO: design a data structure for this, will make code more readable below
	// Also the data structures can incorporate the edge cases with and glue at the end of the line
	// maybe a single data structure for all 3
	// lineData.Width(start int64, end int64)
	lineData := buildLineData(items)

	// Set of nodes
	activeNodes := make(map[node]bool)

	firstNode := node{position: 0, line: 0, fitnessClass: 1}
	activeNodes[firstNode] = true

	nodeToPrevious := make(map[node]*node)
	nodeToMinDemerits := make(map[node]float64)
	for position := int64(0); position < int64(len(items)); position++ {
		if items[position].PenaltyCost() > 10000 {
			continue
		}
		if IsBox(items[position]) {
			continue
		}
		if IsGlue(items[position]) && (position == 0 || !IsBox(items[position-1])) {
			continue
		}
		for activeNode := range activeNodes {
			adjustmentRatio := calculateAdjustmentRatio(
				lineData.GetWidth(activeNode.position, position),
				lineData.GetShrinkability(activeNode.position, position),
				lineData.GetStrechability(activeNode.position, position),
				subsequentLineLengths,
			)

			// We can't break here, and we can't break in any future positions using this
			// node because the adjustmentRatio is only going to get worse
			if adjustmentRatio < -1 {
				delete(activeNodes, activeNode)
				continue
			}
			// We must add a break here, in which case previous active nodes are deleted
			if items[position].PenaltyCost() < -10000 {
				delete(activeNodes, activeNode)
			}
			// This is the case when there is not enough material for the line.
			// We skip, but keep the node active because in a future breakpoint it may be used.
			if adjustmentRatio > criteria.GetMaxAdjustmentRatio() {
				continue
			}
			thisNode := node{
				position:     position,
				line:         activeNode.line + 1, // Todo: should be -1 in some cases
				fitnessClass: criteria.CalculateFitnessClass(adjustmentRatio),
			}
			demerits := criteria.CalculateDemerits(
				adjustmentRatio,
				thisNode.fitnessClass,
				activeNode.fitnessClass,
				items[position].PenaltyCost(),
				items[position].IsFlaggedPenalty(),
				items[position-1].IsFlaggedPenalty(),
			) + nodeToMinDemerits[activeNode]
			minDemeritsSoFar := true
			if nodeToPrevious[thisNode] != nil {
				minDemeritsSoFar = demerits < nodeToMinDemerits[thisNode]
			}
			if minDemeritsSoFar {
				nodeToPrevious[thisNode] = &activeNode
				nodeToMinDemerits[thisNode] = demerits
			}
		}
	}
	if len(activeNodes) == 0 {
		return nil, &NoSolutionError{}
	}
	var bestNode *node
	minDemerits := float64(-1)
	for activeNode := range activeNodes {
		if minDemerits < 0 || nodeToMinDemerits[activeNode] < minDemerits {
			minDemerits = nodeToMinDemerits[activeNode]
			bestNode = &activeNode
		}
	}
	numBreakpoints := int64(0)
	for thisNode := bestNode; thisNode != nil; thisNode = nodeToPrevious[*thisNode] {
		numBreakpoints++
	}
	breakpointIndices := make([]int64, numBreakpoints)
	for thisNode := bestNode; thisNode != nil; thisNode = nodeToPrevious[*thisNode] {
		breakpointIndices[numBreakpoints-1] = thisNode.position
		numBreakpoints--
	}
	return breakpointIndices, nil
}

// lineData is a data structure that facilitates computing the width, shrinkability
// and stretchability of a line (i.e., a contiguous subsequence of items) in O(1).
// It incorporates the logic that the width of a glue item doesn't count if the glue
// is at the end of the line, and that the width of a penalty item only counts if 
// it is at the end.
// TODO: probably this logic should be on the structure itself
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
	for position := int64(0); position < int64(len(items)); position++ {
		lineData.aggregateWidth[position+1] = lineData.aggregateWidth[position]
		if !IsPenalty(items[position]) {
			lineData.aggregateWidth[position+1] += items[position].Width()
		}
		lineData.aggregateShrinkability[position+1] = lineData.aggregateShrinkability[position] + items[position].Shrinkability()
		lineData.aggregateStretchibility[position+1] = lineData.aggregateStretchibility[position] + items[position].Stretchability()
	}
	return lineData
}

func (lineData *lineData) GetWidth(start int64, end int64) (width int64) {
	width = lineData.aggregateWidth[end+1] - lineData.aggregateWidth[start]
	if IsPenalty(lineData.items[end]) {
		width += lineData.items[end].Width()
	} else if IsGlue(lineData.items[end]) {
		width -= lineData.items[end].Width()
	}
	return
}

func (lineData *lineData) GetShrinkability(start int64, end int64) (shrinkability int64) {
	shrinkability = lineData.aggregateShrinkability[end+1] - lineData.aggregateShrinkability[start]
	if IsGlue(lineData.items[end]) {
		shrinkability -= lineData.items[end].Shrinkability()
	}
	return
}
func (lineData *lineData) GetStrechability(start int64, end int64) (stretchibility int64) {
	stretchibility = lineData.aggregateStretchibility[end+1] - lineData.aggregateStretchibility[start]
	if IsGlue(lineData.items[end]) {
		stretchibility -= lineData.items[end].Stretchability()
	}
	return
}

// NoSolutionError is returned if the problem has no solution satisfying the optimality contraints
type NoSolutionError struct{}

func (err *NoSolutionError) Error() string {
	return "No solution!"
}

func calculateAdjustmentRatio(
	lineWidth int64,
	lineShrinkability int64,
	lineStretchability int64,
	targetLineWidth int64,

) float64 {
	if lineWidth < targetLineWidth {
		// Division by 0, inf is allowed
		return float64(lineWidth-targetLineWidth) / float64(lineShrinkability)
	}
	if lineWidth > targetLineWidth {
		return float64(lineWidth-targetLineWidth) / float64(lineStretchability)
	}
	return 0
}

type node struct {
	position     int64
	line         int64
	fitnessClass FitnessClass
	// Note: total width needs to account for penalty width of the breakpoint
}
