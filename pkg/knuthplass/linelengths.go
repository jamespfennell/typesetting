package knuthplass

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
	return len(lineLengths.initialLengths)
}
