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
				CommandToken{"a"},
				CharacterToken{"{", catcode.BeginGroup},
				CharacterToken{"b", catcode.Letter},
				CharacterToken{"}", catcode.EndGroup},
			},
		},
		{
			"\\a b",
			[]Token{
				CommandToken{"a"},
				CharacterToken{"b", catcode.Letter},
			},
		},
		{
			"\\a  b",
			[]Token{
				CommandToken{"a"},
				CharacterToken{" ", catcode.Space},
				CharacterToken{"b", catcode.Letter},
			},
		},
		{
			"\\ABC{D}",
			[]Token{
				CommandToken{"ABC"},
				CharacterToken{"{", catcode.BeginGroup},
				CharacterToken{"D", catcode.Letter},
				CharacterToken{"}", catcode.EndGroup},
			},
		},
		{
			"\\ABC",
			[]Token{
				CommandToken{"ABC"},
			},
		},
		{
			"\\{{",
			[]Token{
				CommandToken{"{"},
				CharacterToken{"{", catcode.BeginGroup},
			},
		},
		{
			"A%a comment here\nC",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CharacterToken{"C", catcode.Letter},
			},
		},
		{
			"A%a comment here\n%A second comment\nC",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CharacterToken{"C", catcode.Letter},
			},
		},
		{
			"A%a comment here",
			[]Token{
				CharacterToken{"A", catcode.Letter},
			},
		},
		{
			"A  B",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CharacterToken{" ", catcode.Space},
				CharacterToken{"B", catcode.Letter},
			},
		},
		{
			"A\nB",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CharacterToken{"\n", catcode.Space},
				CharacterToken{"B", catcode.Letter},
			},
		},
		{
			"A \nB",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CharacterToken{" ", catcode.Space},
				CharacterToken{"B", catcode.Letter},
			},
		},
		{
			"A\n\nB",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CommandToken{"par"},
				CharacterToken{"B", catcode.Letter},
			},
		},
		{
			"A\n \nB",
			[]Token{
				CharacterToken{"A", catcode.Letter},
				CommandToken{"par"},
				CharacterToken{"B", catcode.Letter},
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
		CharacterToken{"B", catcode.Letter},
	}
	verifyAllValidTokens(t, tokenizer, expected)
}

func TestTokenizer_InvalidCharacter(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("B", catcode.Invalid)
	tokenizer := NewTokenizer(strings.NewReader("AB"), &m)
	verifyValidToken(t, tokenizer, CharacterToken{"A", catcode.Letter})
	verifyInvalidToken(t, tokenizer)
}

func TestTokenizer_NonUtf8Character(t *testing.T) {
	paramsList := []string{"A", "A%", "A\\"}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			m := catcode.NewCatCodeMapWithTexDefaults()
			s := params + string([]byte{0b11000010, 0b00100010})
			tokenizer := NewTokenizer(strings.NewReader(s), &m)
			verifyValidToken(t, tokenizer, CharacterToken{"A", catcode.Letter})
			verifyInvalidToken(t, tokenizer)
		})
	}
}

func TestTokenizer_ScopeChange(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	tokenizer := NewTokenizer(strings.NewReader("{{{"), &m)
	verifyValidToken(t, tokenizer, CharacterToken{"{", catcode.BeginGroup})
	m.BeginScope()
	m.Set("{", catcode.Subscript)
	verifyValidToken(t, tokenizer, CharacterToken{"{", catcode.Subscript})
	m.EndScope()
	verifyValidToken(t, tokenizer, CharacterToken{"{", catcode.BeginGroup})
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
	if finalToken != nil {
		t.Fatalf("Expected to recieve nil token last but recieved: %v", finalToken)
	}
	if err != io.EOF {
		t.Fatalf("Expected io.EOF error in retriving last token but recieved: %s", err)
	}
}
