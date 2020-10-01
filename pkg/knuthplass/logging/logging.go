package logging

import (
	"fmt"
	"io"
	"strings"
)

// BreakpointLogger contains data gathered during the Knuth-Plass algorithm.
type BreakpointLogger struct {
	AdjustmentRatiosTable *Table
	LineDemeritsTable     *Table
	TotalDemeritsTable    *Table
}

// NewBreakpointLogger returns a new initialized BreakpointLogger.
func NewBreakpointLogger() *BreakpointLogger {
	return &BreakpointLogger{
		AdjustmentRatiosTable: NewTable("Adjustment ratios"),
		LineDemeritsTable:     NewTable("Line demerits"),
		TotalDemeritsTable:    NewTable("Total demerits"),
	}
}

// Key is an interface for the row/column keys that can be used in a Table.
type Key interface {
	ID() string
}

// DataPoint is an interface for the kind of data that can be placed in a cell of a Table.
type DataPoint interface {
	String() string
}

// Table is a data structure for holding tabular data where both the row and column indices are Keys.
type Table struct {
	name       string
	keyTracker keyTracker
	data       map[string]map[string]DataPoint
}

// AddCell adds a new data point to the Table (or replaces the data point, if it already exists).
func (table *Table) AddCell(row Key, col Key, value DataPoint) {
	table.keyTracker.initNode(row)
	table.keyTracker.initNode(col)
	_, rowExists := table.data[row.ID()]
	if !rowExists {
		table.data[row.ID()] = make(map[string]DataPoint)
	}
	table.data[row.ID()][col.ID()] = value
}

func fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		fmt.Println("Encountered error when printing to string buffer: ", err)
	}
}

// String returns the Table as a string.
func (table *Table) String() string {
	var b strings.Builder
	fprintf(&b, " +---[ %s ]---\n", table.name)
	headerLine := fmt.Sprintf(" | %10s | ", "")

	// Calculate the width of each column
	colKeyToColWidth := make(map[string]int)
	for _, colKey := range table.keyTracker.keysInOrder {
		colKeyToColWidth[colKey.ID()] = -1.0
		for _, rowKey := range table.keyTracker.keysInOrder {
			value, hasValue := table.data[rowKey.ID()][colKey.ID()]
			if !hasValue {
				continue
			}
			stringValue := fmt.Sprintf("%9s", value)
			if len(stringValue) > colKeyToColWidth[colKey.ID()] {
				colKeyToColWidth[colKey.ID()] = len(stringValue)
			}
		}
	}

	// String the header row
	fprintf(&b, " | %10s | ", "")
	for _, colKey := range table.keyTracker.keysInOrder {
		if colKeyToColWidth[colKey.ID()] < 0 {
			continue
		}
		headerLine = headerLine + fmt.Sprintf(" %*s |", colKeyToColWidth[colKey.ID()], colKey.ID())
		fprintf(&b, " %*s |", colKeyToColWidth[colKey.ID()], colKey.ID())
	}
	fprintf(&b, "\n")

	for _, rowKey := range table.keyTracker.keysInOrder {
		// We store this row's data in another buffer. If there is no data in the whole row, we won't print it.
		var lineB strings.Builder
		fprintf(&lineB, " | %10s | ", rowKey.ID())

		rowHasValues := false
		for _, colKey := range table.keyTracker.keysInOrder {
			if colKeyToColWidth[colKey.ID()] < 0 {
				continue
			}
			value, hasValue := table.data[rowKey.ID()][colKey.ID()]
			if hasValue {
				stringValue := fmt.Sprintf("%9s", value)
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey.ID()], stringValue)
				rowHasValues = true
			} else {
				fprintf(&lineB, " %*s |", colKeyToColWidth[colKey.ID()], "")
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

// NewTable creates a new table with the given name.
func NewTable(name string) *Table {
	return &Table{
		name:       name,
		keyTracker: newKeyTracker(),
		data:       make(map[string]map[string]DataPoint),
	}
}

type keyTracker struct {
	keysInOrder []Key
	keysSeen    map[string]bool
}

func newKeyTracker() keyTracker {
	return keyTracker{
		keysInOrder: []Key{},
		keysSeen:    make(map[string]bool),
	}
}
func (tracker *keyTracker) initNode(thisNode Key) {
	if tracker.keysSeen[thisNode.ID()] {
		return
	}
	tracker.keysSeen[thisNode.ID()] = true
	tracker.keysInOrder = append(tracker.keysInOrder, thisNode)
}
