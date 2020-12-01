package stream

import (
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"github.com/jamespfennell/typesetting/pkg/tex/tokenization/catcode"
	"strings"
)

func Consume(stream token.Stream) error {
	for {
		t, err := stream.NextToken()
		if err != nil {
			return err
		}
		if t == nil {
			return nil
		}
	}
}

func ReadString(stream token.Stream) (string, error) {
	var b strings.Builder
	for {
		t, err := stream.PeekToken()
		if err != nil {
			return "", err
		}
		if t.IsCommand() || t.CatCode() == catcode.Space {
			return b.String(), nil
		}
		_, _ = stream.NextToken()
		b.WriteString(t.Value())
	}
}
