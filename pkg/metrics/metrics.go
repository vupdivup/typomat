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
