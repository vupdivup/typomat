// Package random provides utilities for randomization tasks.
package random

import (
	"math/rand/v2"
)

// Shuffle returns a new slice with the elements of the input slice shuffled
// randomly using the Fisher-Yates algorithm.
func Shuffle[T any](slice []T) []T {
	s := make([]T, len(slice))
	copy(s, slice)

	for i := len(s) - 1; i > 0; i-- {
		j := rand.IntN(i + 1)
		s[i], s[j] = s[j], s[i]
	}

	return s
}
