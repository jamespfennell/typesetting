package logging

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
		nodesInOrder: []keyable{},
		nodesSeen:    make(map[string]bool),
	}
	return &BreakpointLogger{
		AdjustmentRatiosTable: newNodeNodeTable("Adjustment ratios", tracker),
		LineDemeritsTable:     newNodeNodeTable("Line demerits", tracker),
		TotalDemeritsTable:    newNodeNodeTable("Total demerits", tracker),
	}
}

type keyable interface {
	Key() string
}

type stringable interface {
	String() string
}

// NodeNodeTable is a data structure for holding tabular data where both the row and column indices are nodes.
type NodeNodeTable struct {
	name        string
	nodeTracker *nodeTracker
	data        map[string]map[string]stringable
}

// AddCell adds a new data point to the NodeNodeTable (or replaces the data point, if it already exists).
func (table *NodeNodeTable) AddCell(rowKey keyable, colKey keyable, value stringable) {
	table.nodeTracker.initNode(rowKey)
	table.nodeTracker.initNode(colKey)
	_, rowExists := table.data[rowKey.Key()]
	if !rowExists {
		table.data[rowKey.Key()] = make(map[string]stringable)
	}
	table.data[rowKey.Key()][colKey.Key()] = value
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
	colKeyToColWidth := make(map[string]int)
	for _, colKey := range table.nodeTracker.nodesInOrder {
		colKeyToColWidth[colKey.Key()] = -1.0
		for _, rowKey := range table.nodeTracker.nodesInOrder {
			value, hasValue := table.data[rowKey.Key()][colKey.Key()]
			if !hasValue {
				continue
			}
			stringValue := fmt.Sprintf("%9s", value)
			if len(stringValue) > colKeyToColWidth[colKey.Key()] {
				colKeyToColWidth[colKey.Key()] = len(stringValue)
			}
		}
	}

	// String the header row
	fprintf(&b, " | %10s | ", "")
	for _, colKey := range table.nodeTracker.nodesInOrder {
		if colKeyToColWidth[colKey.Key()] < 0 {
			continue
		}
		headerLine = headerLine + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey.Key()], colKey.Key())
		fprintf(&b, " %*s |", colKeyToColWidth[colKey.Key()], colKey.Key())
	}
	fprintf(&b, "\n")

	for _, rowKey := range table.nodeTracker.nodesInOrder {
		// We store this row's data in another buffer. If there is no data in the whole row, we won't print it.
		var lineB strings.Builder
		fprintf(&lineB, " | %10s | ", rowKey.Key())

		rowHasValues := false
		for _, colKey := range table.nodeTracker.nodesInOrder {
			if colKeyToColWidth[colKey.Key()] < 0 {
				continue
			}
			value, hasValue := table.data[rowKey.Key()][colKey.Key()]
			if hasValue {
				stringValue := fmt.Sprintf("%9s", value)
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey.Key()], stringValue)
				rowHasValues = true
			} else {
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey.Key()], "")
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
		data:        make(map[string]map[string]stringable),
	}
}

type nodeTracker struct {
	nodesInOrder []keyable
	nodesSeen    map[string]bool
}

func (tracker *nodeTracker) initNode(thisNode keyable) {
	if tracker.nodesSeen[thisNode.Key()] {
		return
	}
	tracker.nodesSeen[thisNode.Key()] = true
	tracker.nodesInOrder = append(tracker.nodesInOrder, thisNode)
}
