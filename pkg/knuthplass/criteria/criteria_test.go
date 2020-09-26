package criteria

import (
	d "github.com/jamespfennell/typesetting/pkg/distance"
	"testing"
)

func TestAllValuesSeen(t *testing.T) {
	allValues := make(map[int64]bool)
	for i := d.Distance(0); i < 10000; i += 3 {
		for j := d.Distance(0); j < 1000; j += 10 {
			allValues[calculateBadness(d.Ratio{Num: i, Den: j})] = true
		}
	}
	// The magic number 1095 comes from Knuth in section 108 of TeX.
	if len(allValues) != 1095 {
		t.Errorf("Expected: %d; Actual: %d", 1095, len(allValues))
	}

}
