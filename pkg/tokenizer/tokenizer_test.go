package tokenizer

import (
	"path/filepath"
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
		// empty string
		{"", []string{}},
		// string with only delimiters
		{" ,;! ", []string{}},
		// numbers and alphanumeric tokens
		{"var1 = 42;", []string{"var"}},
		// acronyms
		{"JSONData", []string{"json", "data"}},
		{"NumCPU", []string{"num", "cpu"}},
	}

	for _, c := range cases {
		got := TokenizeString(c.input, nil)
		assert.ElementsMatch(t, got, c.want)
	}
}

func TestTokenizeFile(t *testing.T) {
	// Expected tokens from all test cases combined
	expected := []string{
		"hello", "world",
		"pascal", "case", "word",
		"camel", "case", "word",
		"snake", "case", "word",
		"var",
		"json", "data",
		"num", "cpu",
	}

	cases := []struct {
		file string
		want []string
	}{
		{"testdata/trailing_newline.txt", expected},
		{"testdata/no_trailing_newline.txt", expected},
	}

	// Test with relative paths
	for _, c := range cases {
		got, err := TokenizeFile(c.file, nil)
		assert.NoError(t, err)
		assert.ElementsMatch(t, got, c.want)
	}

	// Test with absolute paths
	for _, c := range cases {
		abs, err := filepath.Abs(c.file)
		assert.NoError(t, err)
		got, err := TokenizeFile(abs, nil)
		assert.NoError(t, err)
		assert.ElementsMatch(t, got, c.want)
	}
}
