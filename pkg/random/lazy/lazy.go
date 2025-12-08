// Package lazy provides lazy random sampling utilities.
package lazy

import (
	"iter"
	"math/rand/v2"
)

// Sample returns k random elements from the provided iterable sequence using
// reservoir sampling.
func Sample[T any](iter iter.Seq[T], k int) []T {
	sample := make([]T, 0, k)
	i := 0

	for item := range iter {
		if i < k {
			sample = append(sample, item)
		} else {
			j := rand.IntN(i + 1)
			if j < k {
				sample[j] = item
			}
		}
		i++
	}

	return sample
}