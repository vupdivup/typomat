// metrics calculates various typing metrics.
package metrics

import (
	"time"
)

// WPM calculates the words per minute (WPM) based on the input string and the
// time duration provided.
// A word is defined as five characters.
func WPM(input string, time time.Duration) float64 {
	numChars := 0

	for _, r := range input {
		if r != ' ' && r != '\n' && r != '\t' {
			numChars++
		}
	}

	minutes := time.Minutes()

	return float64(numChars) / 5.0 / minutes
}

// Accuracy calculates the typing accuracy as a percentage based on the prompt
// and the user's input.
func Accuracy(prompt, input string) float64 {
	if len(prompt) == 0 {
		return 100.0
	}

	numCorrect := 0

	for i, r := range prompt {
		if i < len(input) && r == rune(input[i]) {
			numCorrect++
		}
	}

	return float64(numCorrect) / float64(len(prompt)) * 100.0
}
