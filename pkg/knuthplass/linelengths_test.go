package knuthplass

import "testing"

func TestConstantLineLengths(t *testing.T) {
	lineLengths := NewConstantLineLengths(2000)

	for i := 0; i < 100; i++ {
		if lineLengths.GetLength(i) != 2000 {
			t.Errorf("Length of line %d is not correct", i)
		}
		if lineLengths.GetNextPseudoIndex(i) != 0 {
			t.Errorf(
				"Actual pseudo index %d is not expected pseudo index %d",
				lineLengths.GetNextPseudoIndex(i),
				0,
			)
		}
	}
}

func TestVariableLineLengths(t *testing.T) {
	expectedLineLengths := []int64{3000, 1000, 4000}
	expectedNextPseudoIndex := []int{1, 2, 3}
	lineLengths := NewVariableLineLengths([]int64{3000, 1000, 4000}, 2000)

	for i := 0; i < 3; i++ {
		if lineLengths.GetLength(i) != expectedLineLengths[i] {
			t.Errorf("Length of line %d is not correct", i)
		}
		if lineLengths.GetNextPseudoIndex(i) != expectedNextPseudoIndex[i] {
			t.Errorf(
				"Actual pseudo index %d is not expected pseudo index %d",
				lineLengths.GetNextPseudoIndex(i),
				expectedNextPseudoIndex[i],
			)
		}
	}
	for i := 3; i < 100; i++ {
		if lineLengths.GetLength(i) != 2000 {
			t.Errorf("Length of line %d is not correct", i)
		}
		if lineLengths.GetNextPseudoIndex(i) != 3 {
			t.Errorf(
				"Actual pseudo index %d is not expected pseudo index %d",
				lineLengths.GetNextPseudoIndex(i),
				3,
			)
		}
	}
}
