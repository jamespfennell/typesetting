package knuthplass

import "fmt"

func NewBreakpointLogger() *BreakpointLogger {
	tracker := &nodeTracker{
		nodesInOrder: []node{},
		nodesSeen:    make(map[node]bool),
	}
	return &BreakpointLogger{
		AdjustmentRatiosTable: NewNodeNodeTable("Adjustment ratios", tracker),
		LineDemeritsTable:     NewNodeNodeTable("Line demerits", tracker),
		TotalDemeritsTable:    NewNodeNodeTable("Total demerits", tracker),
	}
}

type BreakpointLogger struct {
	AdjustmentRatiosTable *NodeNodeTable
	LineDemeritsTable     *NodeNodeTable
	TotalDemeritsTable    *NodeNodeTable
}

type nodeTracker struct {
	nodesInOrder []node
	nodesSeen    map[node]bool
}

func (tracker *nodeTracker) initNode(thisNode node) {
	if _, nodeSeen := tracker.nodesSeen[thisNode]; nodeSeen {
		return
	}
	tracker.nodesSeen[thisNode] = true
	tracker.nodesInOrder = append(tracker.nodesInOrder, thisNode)
}

func NewNodeNodeTable(name string, tracker *nodeTracker) *NodeNodeTable {
	return &NodeNodeTable{
		name:        name,
		nodeTracker: tracker,
		data:        make(map[node]map[node]float64),
	}
}

type NodeNodeTable struct {
	name        string
	nodeTracker *nodeTracker
	data        map[node]map[node]float64
}

func (table *NodeNodeTable) AddCell(rowKey node, colKey node, value float64) {
	table.nodeTracker.initNode(rowKey)
	table.nodeTracker.initNode(colKey)
	_, rowExists := table.data[rowKey]
	if !rowExists {
		table.data[rowKey] = make(map[node]float64)
	}
	table.data[rowKey][colKey] = value
}

func (table *NodeNodeTable) Print() {
	fmt.Println(fmt.Sprintf(" +---[ %s ]---", table.name))
	headerLine := fmt.Sprintf(" | %10s | ", "")

	colKeyToColWidth := make(map[node]int)
	for _, colKey := range table.nodeTracker.nodesInOrder {
		colKeyToColWidth[colKey] = -1.0
		for _, rowKey := range table.nodeTracker.nodesInOrder {
			value, hasValue := table.data[rowKey][colKey]
			if !hasValue {
				continue
			}
			stringValue := fmt.Sprintf("%6.2f", value)
			if len(stringValue) > colKeyToColWidth[colKey] {
				colKeyToColWidth[colKey] = len(stringValue)
			}
		}
	}

	for _, colKey := range table.nodeTracker.nodesInOrder {
		if colKeyToColWidth[colKey] < 0 {
			continue
		}
		headerLine = headerLine + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey], buildNodeLabel(colKey))
	}

	fmt.Println(headerLine)
	for _, rowKey := range table.nodeTracker.nodesInOrder {
		line := fmt.Sprintf(" | %10s | ", buildNodeLabel(rowKey))
		rowHasValues := false
		for _, colKey := range table.nodeTracker.nodesInOrder {
			if colKeyToColWidth[colKey] < 0 {
				continue
			}
			value, hasValue := table.data[rowKey][colKey]
			if hasValue {
				stringValue := fmt.Sprintf("%6.2f", value)
				line = line + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey], stringValue)
				rowHasValues = true
			} else {
				line = line + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey], "")
			}
		}
		if !rowHasValues {
			continue
		}
		fmt.Println(line)
	}
	fmt.Println(" +----")
	fmt.Println(" Node key: [item index]/[line index]/[fitness class]. Line index may...")
	fmt.Println()
}

func buildNodeLabel(node node) string {
	return fmt.Sprintf("%d/%d/%d", node.itemIndex, node.lineIndex, node.fitnessClass)
}
