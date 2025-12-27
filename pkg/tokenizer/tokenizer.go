// tokenizer can be used to split strings into lexical tokens.
package tokenizer

import (
	"bufio"
	"io"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/vupdivup/typomat/pkg/alphabet"
)

type Case int

const (
	CaseLower Case = iota
	CaseUpper
)

var (
	wordDelimiters = []rune{
		// ASCII Whitespace: space, tab, newline, vertical tab, form feed,
		// carriage return
		' ', '\t', '\n', '\v', '\f', '\r',
	}
	tokenDelimiters = []rune{
		// Punctuation marks used in text and code
		'.', ',', '!', '?', ';', ':', '(', ')', '[', ']', '{', '}', '"', '\'',
		'-', '_', '/', '<', '>', '@', '`',

		// Alphanumeric characters
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	}
)

func getRuneCase(r rune) Case {
	if unicode.IsUpper(r) {
		return CaseUpper
	}
	return CaseLower
}

// splitMixedCaseToken splits a mixed-case identifier into its subtokens.
// For example, "CamelCaseWord" becomes "Camel", "Case", "Word".
// Acronyms are treated as single tokens, e.g., "JSONData" becomes "JSON",
// "Data".
// No lowercase conversion is performed here.
func splitMixedCaseToken(word []rune) [][]rune {
	var result [][]rune
	var currentToken []rune

	flush := func() {
		if len(currentToken) > 0 {
			result = append(result, currentToken)
			currentToken = []rune{}
		}
	}

	for i, r := range word {
		// flush if:
		// 1) transition from lower (previous) to upper case (current)
		// 2) transition from upper (current) to lower case (next)
		isPreviousLower := i > 0 && getRuneCase(word[i-1]) == CaseLower
		isNextLower := i < len(word)-1 && getRuneCase(word[i+1]) == CaseLower
		if getRuneCase(r) == CaseUpper && (isPreviousLower || isNextLower) {
			flush()
		}

		currentToken = append(currentToken, r)
	}

	flush()
	return result
}

// isTokenValid checks if all runes in the token are part of the allowed
// character set.
func isTokenValid(token []rune) bool {
	for _, r := range token {
		if !slices.Contains(alphabet.AllRunes, r) {
			return false
		}
	}
	return true
}

func tokenizeWord(word []rune) []string {
	var tokens []string
	var currentToken []rune

	flush := func() {
		if len(currentToken) > 0 && isTokenValid(currentToken) {
			subtokens := splitMixedCaseToken(currentToken)
			for _, subtoken := range subtokens {
				if isTokenValid(subtoken) {
					tokens = append(tokens, strings.ToLower(string(subtoken)))
				}
			}
		}

		currentToken = []rune{}
	}

	for _, char := range word {
		if slices.Contains(tokenDelimiters, char) {
			flush()
			continue
		}

		currentToken = append(currentToken, char)
	}

	flush()
	return tokens
}

// TokenizeString splits an input string into word tokens.
// It can handle natural language text as well as source code.
// Tokens containing non-ASCII characters are filtered out.
//
// An optional filter function can be provided to include/exclude specific words
// before tokenization. It should return true to include the word.
func TokenizeString(s string, wordFilter func(string) bool) []string {
	var tokens []string
	var currentWord []rune

	if wordFilter == nil {
		wordFilter = func(string) bool { return true }
	}

	flush := func() {
		// Check word filter
		if wordFilter(string(currentWord)) {
			tokens = append(tokens, tokenizeWord(currentWord)...)
		}
		currentWord = []rune{}
	}

	for _, r := range s {
		// Check for token boundaries
		if slices.Contains(wordDelimiters, r) {
			flush()
			continue
		}

		currentWord = append(currentWord, r)
	}

	flush()
	return tokens
}

// TokenizeFile reads a file and returns its tokens. File contents are read
// line by line to handle large files efficiently.
// See TokenizeString for tokenization details.
func TokenizeFile(path string, wordFilter func(string) bool) ([]string, error) {
	tokens := []string{}

	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return tokens, err
	}
	defer file.Close()

	reader := bufio.NewReaderSize(file, 128*1024)
	for {
		// Read and tokenize each line
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return tokens, err
		}
		lineTokens := TokenizeString(line, wordFilter)
		tokens = append(tokens, lineTokens...)
	}

	return tokens, nil
}
