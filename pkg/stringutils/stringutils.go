// Package stringutils provides utility functions for string manipulation.
package stringutils

import (
	"strings"
)

// Wrap splits the input string into lines not exceeding the specified line
// width, using a custom word separator rune.
func Wrap(s string, lineWidth int, space rune) []string {
	if space == '\x00' {
		space = ' '
	}

	words := strings.Split(s, string(space))
	lines := []string{}
	line := []string{}
	lineLen := 0

	flush := func() {
		if len(line) == 0 {
			return
		}
		lines = append(lines, strings.Join(line, string(space)))
		line = []string{}
		lineLen = 0
	}

	for _, word := range words {
		wordRunes := []rune(word)

		if lineLen > 0 && lineLen + len(wordRunes) > lineWidth {
			flush()
		}

		line = append(line, word)
		lineLen += len(wordRunes) + 1
	}

	flush()
	return lines
}
