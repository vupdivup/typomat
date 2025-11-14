// tokenizer can be used to split strings into lexical tokens.
package tokenizer

import (
	"slices"
)

var tokenDelimiters = []rune{
	// ASCII Whitespace: space, tab, newline, vertical tab, form feed,
	// carriage return
	' ', '\t', '\n', '\v', '\f', '\r',

	// Punctuation marks used in text and code
	'.', ',', '!', '?', ';', ':', '(', ')', '[', ']', '{', '}', '"', '\'', '-',
	'_',

	// Alphanumeric characters
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
}

var alphabetRunes = []rune{
	'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm',
	'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M',
	'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
}

// isTokenValid checks if all runes in the token are part of the allowed
// character set.
func isTokenValid(token []rune) bool {
	for _, r := range token {
		if !slices.Contains(alphabetRunes, r) {
			return false
		}
	}
	return true
}

// Tokenize splits an input string into word tokens.
// It can handle natural language text as well as source code.
// Tokens containing non-ASCII characters are filtered out.
func Tokenize(s string) []string { // TODO: lazy iterator
	var tokens []string
	var currentToken []rune

	flush := func() {
		if len(currentToken) > 0 && isTokenValid(currentToken) {
			tokens = append(tokens, string(currentToken))
		}
				
		currentToken = []rune{}
	}

	for _, r := range s {
		// Check for token boundaries
		if slices.Contains(tokenDelimiters, r) {
			flush()
			continue
		}

		currentToken = append(currentToken, r)
	}

	flush()
	return tokens
}
