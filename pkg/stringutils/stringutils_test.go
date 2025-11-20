package stringutils

import (
	"testing"
)

func TestWrap(t *testing.T) {
	cases := []struct {
		input     string
		lineWidth int
		space rune
		want []string
	}{
		{"abc def ghi", 10, ' ', []string{"abc def", "ghi"}},
		{"abc def ghijkl", 7, ' ', []string{"abc def", "ghijkl"}},
		{"ábcdéf ghi", 8, ' ', []string{"ábcdéf", "ghi"}},
		{"word1,word2,word3", 11, ',', []string{"word1,word2", "word3"}},
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