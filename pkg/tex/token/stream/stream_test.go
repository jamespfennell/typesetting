package stream_test

import (
	"github.com/jamespfennell/typesetting/pkg/tex/testutil"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"testing"
)

func TestChainedStream_MultipleStream(t *testing.T) {
	s1 := testutil.NewSimpleStream("a", "b", "c")
	s2 := testutil.NewSimpleStream("d", "e", "f")
	s := stream.NewChainedStream(s1, s2)
	output := testutil.NewSimpleStream("a", "b", "c", "d", "e", "f")
	testutil.CheckStreamEqual(t, s, output)
}
