package tokenizer

import (
	"bufio"
	"io"
	"slices"
)

// Space, tab, newline, vertical tab, form feed, carriage return
var asciiWhitespaceRunes = []rune{' ', '\t', '\n', '\v', '\f', '\r'}

// TokenizeStream reads from a bufio.Reader and tokenizes the input based on
// ASCII whitespace.
func TokenizeStream(b *bufio.Reader) ([]string, error) {  // TODO: lazy iterator
	var tokens []string
	var currentToken []rune

	flush := func() {
		if len(currentToken) > 0 {
			tokens = append(tokens, string(currentToken))
			currentToken = []rune{}
		}
	}

	for {
		r, _, err := b.ReadRune()

		if err == io.EOF {
			break
		} else if err != nil {
			return []string{}, err
		}

		if slices.Contains(asciiWhitespaceRunes, r) {
			flush()
			continue
		}
		
		currentToken = append(currentToken, r)
	}

	flush()
	return tokens, nil
}
