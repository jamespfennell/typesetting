package input

import (
	"github.com/jamespfennell/typesetting/pkg/tex/catcode"
	"github.com/jamespfennell/typesetting/pkg/tex/token"
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	paramsList := []struct {
		input          string
		expectedTokens []token.Token
	}{
		{
			"\\a{b}",
			[]token.Token{
				token.NewCommandToken("a", nil),
				token.NewCharacterToken("{", catcode.BeginGroup, nil),
				token.NewCharacterToken("b", catcode.Letter, nil),
				token.NewCharacterToken("}", catcode.EndGroup, nil),
			},
		},
		{
			"\\a b",
			[]token.Token{
				token.NewCommandToken("a", nil),
				token.NewCharacterToken("b", catcode.Letter, nil),
			},
		},
		{
			"\\a  b",
			[]token.Token{
				token.NewCommandToken("a", nil),
				token.NewCharacterToken("b", catcode.Letter, nil),
			},
		},
		{
			"\\a\n b",
			[]token.Token{
				token.NewCommandToken("a", nil),
				token.NewCharacterToken("b", catcode.Letter, nil),
			},
		},
		{
			"\\ABC{D}",
			[]token.Token{
				token.NewCommandToken("ABC", nil),
				token.NewCharacterToken("{", catcode.BeginGroup, nil),
				token.NewCharacterToken("D", catcode.Letter, nil),
				token.NewCharacterToken("}", catcode.EndGroup, nil),
			},
		},
		{
			"\\ABC",
			[]token.Token{
				token.NewCommandToken("ABC", nil),
			},
		},
		{
			"\\{{",
			[]token.Token{
				token.NewCommandToken("{", nil),
				token.NewCharacterToken("{", catcode.BeginGroup, nil),
			},
		},
		{
			"A%a comment here\nC",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken("C", catcode.Letter, nil),
			},
		},
		{
			"A%a comment here\n%A second comment\nC",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken("C", catcode.Letter, nil),
			},
		},
		{
			"A%a comment here",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
			},
		},
		{
			"A%\n B",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A%\n\n B",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCommandToken("par", nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"\\A %\nB",
			[]token.Token{
				token.NewCommandToken("A", nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A  B",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken("  ", catcode.Space, nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A\nB",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken("\n", catcode.Space, nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A \nB",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCharacterToken(" \n", catcode.Space, nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A\n\nB",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCommandToken("par", nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
		{
			"A\n \nB",
			[]token.Token{
				token.NewCharacterToken("A", catcode.Letter, nil),
				token.NewCommandToken("par", nil),
				token.NewCharacterToken("B", catcode.Letter, nil),
			},
		},
	}
	for _, params := range paramsList {
		t.Run(params.input, func(t *testing.T) {
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
	expected := []token.Token{
		token.NewCharacterToken("B", catcode.Letter, nil),
	}
	verifyAllValidTokens(t, tokenizer, expected)
}

func TestTokenizer_IgnoredCharacterInCommandIsAllowed(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("A", catcode.Ignored)
	tokenizer := NewTokenizer(strings.NewReader("\\A"), &m)
	expected := []token.Token{
		token.NewCommandToken("A", nil),
	}
	verifyAllValidTokens(t, tokenizer, expected)
}

func TestTokenizer_InvalidCharacter(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("B", catcode.Invalid)
	tokenizer := NewTokenizer(strings.NewReader("AB"), &m)
	verifyValidToken(t, tokenizer, token.NewCharacterToken("A", catcode.Letter, nil))
	verifyInvalidToken(t, tokenizer)
}

func TestTokenizer_InvalidCharacterInCommentIsAllowed(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("B", catcode.Invalid)
	tokenizer := NewTokenizer(strings.NewReader("A%B"), &m)
	verifyAllValidTokens(t, tokenizer, []token.Token{token.NewCharacterToken("A", catcode.Letter, nil)})
}

func TestTokenizer_InvalidCharacterInCommandIsAllowed(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	m.Set("B", catcode.Invalid)
	tokenizer := NewTokenizer(strings.NewReader("\\B"), &m)
	verifyAllValidTokens(t, tokenizer, []token.Token{token.NewCommandToken("B", nil)})
}

func TestTokenizer_NonUtf8Character(t *testing.T) {
	paramsList := []string{"A", "A%", "A\\"}
	for _, params := range paramsList {
		t.Run("", func(t *testing.T) {
			m := catcode.NewCatCodeMapWithTexDefaults()
			s := params + string([]byte{0b11000010, 0b00100010})
			tokenizer := NewTokenizer(strings.NewReader(s), &m)
			verifyValidToken(t, tokenizer, token.NewCharacterToken("A", catcode.Letter, nil))
			verifyInvalidToken(t, tokenizer)
		})
	}
}

func TestTokenizer_ScopeChange(t *testing.T) {
	m := catcode.NewCatCodeMapWithTexDefaults()
	tokenizer := NewTokenizer(strings.NewReader("{{{"), &m)
	verifyValidToken(t, tokenizer, token.NewCharacterToken("{", catcode.BeginGroup, nil))
	m.BeginScope()
	m.Set("{", catcode.Subscript)
	verifyValidToken(t, tokenizer, token.NewCharacterToken("{", catcode.Subscript, nil))
	m.EndScope()
	verifyValidToken(t, tokenizer, token.NewCharacterToken("{", catcode.BeginGroup, nil))
}

func verifyAllValidTokens(t *testing.T, tokenizer *Tokenizer, expectedTokens []token.Token) {
	for i, _ := range expectedTokens {
		verifyValidToken(t, tokenizer, expectedTokens[i])
	}
	verifyFinalToken(t, tokenizer)
}

func verifyValidToken(t *testing.T, tokenizer *Tokenizer, expectedToken token.Token) {
	actualToken, err := tokenizer.NextToken()
	if err != nil {
		t.Fatalf("Expected no error in retriving token %v but recieved: %s", expectedToken, err)
	}
	if expectedToken.Value() != actualToken.Value() || expectedToken.CatCode() != actualToken.CatCode() {
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
	if err != nil {
		t.Fatalf("Expected no error in retriving last token but recieved: %s", err)
	}
	if finalToken != nil {
		t.Fatalf("Expected to recieve an nil token last but recieved: %v", finalToken)
	}

}
