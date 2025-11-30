package textutils

import (
	"testing"
)

func TestWrap(t *testing.T) {
	cases := []struct {
		input     string
		lineWidth int
		space     rune
		want      []string
	}{
		// basic wrapping without exceeding width
		{"abc def ghi", 10, ' ', []string{"abc def ", "ghi"}},

		// wrap after second word due to narrow width
		{"abc def ghijkl", 7, ' ', []string{"abc ", "def ", "ghijkl"}},

		// multibyte (unicode) characters counted as runes
		{"ábcdéf ghi", 8, ' ', []string{"ábcdéf ", "ghi"}},

		// custom comma separator preserved at line end
		{"word1,word2,word3", 11, ',', []string{"word1,", "word2,word3"}},

		// empty input yields no lines
		{"", 10, ' ', []string{}},

		// single short word under width stays whole
		{"singleword", 20, ' ', []string{"singleword"}},

		// single word exactly equal to width
		{"exactwidth", 10, ' ', []string{"exactwidth"}},

		// very long single word (no internal separators) not split
		{"superlongword", 5, ' ', []string{"superlongword"}},

		// consecutive separators are preserved verbatim
		{"a  b", 10, ' ', []string{"a  b"}},

		// zero rune defaults to space separator
		{"a b", 10, '\x00', []string{"a b"}},

		// wrapping triggered before adding full next word
		{"ab cd ef", 5, ' ', []string{"ab ", "cd ef"}},

		// minimal width forces each word (with separator) on its own line
		{"a b", 1, ' ', []string{"a ", "b"}},

		// trailing separator retained on final line
		{"abc def ", 10, ' ', []string{"abc def "}},

		// comma separator with earlier wrap boundary
		{"one,two,three", 9, ',', []string{"one,two,", "three"}},

		// narrow width forces each word plus separator individually
		{"one two three", 4, ' ', []string{"one ", "two ", "three"}},
	}

	fail := func(input string, lineWidth int, space rune, want, got []string) {
		t.Errorf(
			"Wrap(%q, %d, %q) = %q; want %q",
			input, lineWidth, space, got, want,
		)
	}

	for _, c := range cases {
		got := Wrap(c.input, c.lineWidth, c.space)
		if len(got) != len(c.want) {
			fail(c.input, c.lineWidth, c.space, c.want, got)
			continue
		}

		for i := range got {
			if got[i] != c.want[i] {
				fail(c.input, c.lineWidth, c.space, c.want, got)
				break
			}
		}
	}
}
