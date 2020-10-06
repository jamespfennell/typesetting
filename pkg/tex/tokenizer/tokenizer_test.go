package tokenizer

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"io"
	"strings"
	"testing"
)

// What about "\A  "?

func TestTokenizer(t *testing.T) {
	paramsList := []struct {
		input          string
		expectedTokens []Token
	}{
		{
			"\\a{b}",
			[]Token{
				NewCommandToken("a"),
				NewCharacterToken("{", catcode.BeginGroup),
				NewCharacterToken("b", catcode.Letter),
				NewCharacterToken("}", catcode.EndGroup),
			},
		},
		{
			"\\a b",
			[]Token{
				NewCommandToken("a"),
				NewCharacterToken("b", catcode.Letter),
			},
		},
		{
			"\\a  b",
			[]Token{
				NewCommandToken("a"),
				NewCharacterToken("b", catcode.Letter),
			},
		},
		{
			"\\a\n b",
			[]Token{
				NewCommandToken("a"),
				NewCharacterToken("b", catcode.Letter),
			},
		},
		{
			"\\ABC{D}",
			[]Token{
				NewCommandToken("ABC"),
				NewCharacterToken("{", catcode.BeginGroup),
				NewCharacterToken("D", catcode.Letter),
				NewCharacterToken("}", catcode.EndGroup),
			},
		},
		{
			"\\ABC",
			[]Token{
				NewCommandToken("ABC"),
			},
		},
		{
			"\\{{",
			[]Token{
				NewCommandToken("{"),
				NewCharacterToken("{", catcode.BeginGroup),
			},
		},
		{
			"A%a comment here\nC",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken("C", catcode.Letter),
			},
		},
		{
			"A%a comment here\n%A second comment\nC",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken("C", catcode.Letter),
			},
		},
		{
			"A%a comment here",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
			},
		},
		{
			"A%\n B",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A%\n\n B",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCommandToken("par"),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"\\A %\nB",
			[]Token{
				NewCommandToken("A"),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A  B",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken(" ", catcode.Space),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A\nB",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken("\n", catcode.Space),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A \nB",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCharacterToken(" ", catcode.Space),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A\n\nB",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCommandToken("par"),
				NewCharacterToken("B", catcode.Letter),
			},
		},
		{
			"A\n \nB",
			[]Token{
				NewCharacterToken("A", catcode.Letter),
				NewCommandToken("par"),
				NewCharacterToken("B", catcode.Letter),
			},
		},
	}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			m := catcode.NewCatCodeMapWithTexDefaults()
			tokenizer := NewTokenizer(strings.NewReader(params.input), &m)
			verifyAllValidTokens(t, tokenizer, params.expectedTokens)
		})
	}
}

func TestTokenizer_IgnoredCharacter(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("A", catcode.Ignored)
	tokenizer := NewTokenizer(strings.NewReader("AB"), &m)
	expected := []Token{
		NewCharacterToken("B", catcode.Letter),
	}
	verifyAllValidTokens(t, tokenizer, expected)
}

func TestTokenizer_InvalidCharacter(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("B", catcode.Invalid)
	tokenizer := NewTokenizer(strings.NewReader("AB"), &m)
	verifyValidToken(t, tokenizer, NewCharacterToken("A", catcode.Letter))
	verifyInvalidToken(t, tokenizer)
}

func TestTokenizer_NonUtf8Character(t *testing.T) {
	paramsList := []string{"A", "A%", "A\\"}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			m := catcode.NewCatCodeMapWithTexDefaults()
			s := params + string([]byte{0b11000010, 0b00100010})
			tokenizer := NewTokenizer(strings.NewReader(s), &m)
			verifyValidToken(t, tokenizer, NewCharacterToken("A", catcode.Letter))
			verifyInvalidToken(t, tokenizer)
		})
	}
}

func TestTokenizer_ScopeChange(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	tokenizer := NewTokenizer(strings.NewReader("{{{"), &m)
	verifyValidToken(t, tokenizer, NewCharacterToken("{", catcode.BeginGroup))
	m.BeginScope()
	m.Set("{", catcode.Subscript)
	verifyValidToken(t, tokenizer, NewCharacterToken("{", catcode.Subscript))
	m.EndScope()
	verifyValidToken(t, tokenizer, NewCharacterToken("{", catcode.BeginGroup))
}

func verifyAllValidTokens(t *testing.T, tokenizer *Tokenizer, expectedTokens []Token) {
	for i, _ := range expectedTokens {
		verifyValidToken(t, tokenizer, expectedTokens[i])
	}
	verifyFinalToken(t, tokenizer)
}

func verifyValidToken(t *testing.T, tokenizer *Tokenizer, expectedToken Token) {
	actualToken, err := tokenizer.NextToken()
	if err != nil {
		t.Fatalf("Expected no error in retriving token %v but recieved: %s", expectedToken, err)
	}
	if expectedToken != actualToken {
		t.Fatalf("Expected token: %v; actual token: %v", expectedToken, actualToken)
	}
}

func verifyInvalidToken(t *testing.T, tokenizer *Tokenizer) {
	actualToken, err := tokenizer.NextToken()
	if err == nil {
		t.Fatalf("Expected error in retriving token but recieved none. The token is %v", actualToken)
	}
}

func verifyFinalToken(t *testing.T, tokenizer *Tokenizer) {
	finalToken, err := tokenizer.NextToken()
	if !finalToken.IsNull() {
		t.Fatalf("Expected to recieve an null token last but recieved: %v", finalToken)
	}
	if err != io.EOF {
		t.Fatalf("Expected io.EOF error in retriving last token but recieved: %s", err)
	}
}
