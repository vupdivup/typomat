package tokenizer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokenizeString(t *testing.T) {
	cases := []struct {
		input string
		want  []string
	}{
		// simple sentence with spaces and punctuation
		{"Hello, world!", []string{"hello", "world"}},
		// mixed case words
		{"PascalCaseWord", []string{"pascal", "case", "word"}},
		{"camelCaseWord", []string{"camel", "case", "word"}},
		// snake_case and kebab-case
		{"snake_case-word", []string{"snake", "case", "word"}},
		// numbers and alphanumeric tokens
		{"var1 = 42;", []string{"var"}},
		// acronyms
		{"JSONData", []string{"json", "data"}},
		{"NumCPU", []string{"num", "cpu"}},
		// empty string
		{"", []string{}},
		// string with only delimiters
		{" ,;! ", []string{}},
	}

	for _, c := range cases {
		got := TokenizeString(c.input, nil)
		assert.ElementsMatch(t, got, c.want)
	}
}
