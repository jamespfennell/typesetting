package testutil

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/token/stream"
	"testing"
)

func NewStream(values ...string) stream.TokenStream {
	var tokens []token.Token
	for _, value := range values {
		if len(value) >= 4 && value[:4] == "func" {
			tokens = append(tokens, token.NewCommandToken(value, nil))
		} else {
			tokens = append(tokens, token.NewCharacterToken(value, catcode.Letter, nil))
		}
	}
	return stream.NewSliceStream(tokens)
}

func CheckStreamEqual(t *testing.T, s1, s2 stream.TokenStream) (result bool) {
	result = true
	for {
		t1, err1 := s1.NextToken()
		t2, err2 := s2.NextToken()
		if err1 != err2 || t1 != t2 {
			t.Errorf("Difference! (%v, %v) != (%v, %v)", t1, err1, t2, err2)
			result = false
		}
		if err1 != nil || t1 == nil {
			return
		}
	}
}

