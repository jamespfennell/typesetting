package knuthplass

import (
	"fmt"
	"io"
	"strings"
)

// BreakpointLogger contains data gathered during the Knuth-Plass algorithm.
type BreakpointLogger struct {
	AdjustmentRatiosTable *NodeNodeTable
	LineDemeritsTable     *NodeNodeTable
	TotalDemeritsTable    *NodeNodeTable
}

// NewBreakpointLogger returns a new initialized BreakpointLogger.
func NewBreakpointLogger() *BreakpointLogger {
	tracker := &nodeTracker{
		nodesInOrder: []nodeID{},
		nodesSeen:    make(map[nodeID]bool),
	}
	return &BreakpointLogger{
		AdjustmentRatiosTable: newNodeNodeTable("Adjustment ratios", tracker),
		LineDemeritsTable:     newNodeNodeTable("Line demerits", tracker),
		TotalDemeritsTable:    newNodeNodeTable("Total demerits", tracker),
	}
}

// NodeNodeTable is a data structure for holding tabular data where both the row and column indices are nodes.
type NodeNodeTable struct {
	name        string
	nodeTracker *nodeTracker
	data        map[nodeID]map[nodeID]float64
}

// AddCell adds a new data point to the NodeNodeTable (or replaces the data point, if it already exists).
func (table *NodeNodeTable) AddCell(rowKey *node, colKey nodeID, value float64) {
	table.nodeTracker.initNode(rowKey.id())
	table.nodeTracker.initNode(colKey)
	_, rowExists := table.data[rowKey.id()]
	if !rowExists {
		table.data[rowKey.id()] = make(map[nodeID]float64)
	}
	table.data[rowKey.id()][colKey] = value
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		fmt.Println("Encountered error when printing to string buffer: ", err)
	}
}

func (table *NodeNodeTable) String() string {
	var b strings.Builder
	fprintf(&b, " +---[ %s ]---\n", table.name)
	headerLine := fmt.Sprintf(" | %10s | ", "")

	// Calculate the width of each column
	colKeyToColWidth := make(map[nodeID]int)
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

	// String the header row
	fprintf(&b, " | %10s | ", "")
	for _, colKey := range table.nodeTracker.nodesInOrder {
		if colKeyToColWidth[colKey] < 0 {
			continue
		}
		headerLine = headerLine + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey], buildNodeLabel(colKey))
		fprintf(&b, " %*s |", colKeyToColWidth[colKey], buildNodeLabel(colKey))
	}
	fprintf(&b, "\n")

	for _, rowKey := range table.nodeTracker.nodesInOrder {
		// We store this row's data in another buffer. If there is no data in the whole row, we won't print it.
		var lineB strings.Builder
		fprintf(&lineB, " | %10s | ", buildNodeLabel(rowKey))

		rowHasValues := false
		for _, colKey := range table.nodeTracker.nodesInOrder {
			if colKeyToColWidth[colKey] < 0 {
				continue
			}
			value, hasValue := table.data[rowKey][colKey]
			if hasValue {
				stringValue := fmt.Sprintf("%6.2f", value)
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey], stringValue)
				rowHasValues = true
			} else {
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey], "")
			}
		}
		if !rowHasValues {
			continue
		}
		fprintf(&b, lineB.String())
		fprintf(&b, "\n")
	}
	fprintf(&b, " +----")
	fprintf(&b, " Node id: [item index]/[line index]/[fitness class]. Line index may...")
	fprintf(&b, "\n")
	return b.String()
}

func newNodeNodeTable(name string, tracker *nodeTracker) *NodeNodeTable {
	return &NodeNodeTable{
		name:        name,
		nodeTracker: tracker,
		data:        make(map[nodeID]map[nodeID]float64),
	}
}

func buildNodeLabel(node nodeID) string {
	return fmt.Sprintf("%d/%d/%d", node.itemIndex, node.lineIndex, node.fitnessClass)
}

type nodeTracker struct {
	nodesInOrder []nodeID
	nodesSeen    map[nodeID]bool
}

func (tracker *nodeTracker) initNode(thisNode nodeID) {
	if tracker.nodesSeen[thisNode] {
		return
	}
	tracker.nodesSeen[thisNode] = true
	tracker.nodesInOrder = append(tracker.nodesInOrder, thisNode)
}
