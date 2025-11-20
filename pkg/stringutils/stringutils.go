// Package stringutils provides utility functions for string manipulation.
package stringutils

// Wrap splits the input string into lines not exceeding the specified line
// width, using a custom word separator rune.
// Word separators stick to the end of the preceding word.
func Wrap(s string, lineWidth int, space rune) []string {
	if space == '\x00' {
		space = ' '
	}

	runes := []rune(s)
	lines := []string{}
	curLine := ""
	curLineLen := 0
	wordStart := 0

	for i, r := range runes {
		if r == space || i == len(runes) - 1 {
			// Get the current word including the separator
			word := runes[wordStart:i + 1]

			if curLine != "" && curLineLen + len(word) > lineWidth {
				// Current line full, start new line
				lines = append(lines, curLine)
				curLine = string(word)
				curLineLen = len(word)
			} else {
				// Add word to current line
				curLine += string(word)
				curLineLen += len(word)
			}
			wordStart = i + 1
		}
	}

	if curLine != "" {
		lines = append(lines, curLine)
	}

	return lines
}