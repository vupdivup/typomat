package lazy

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSample(t *testing.T) {
	pop := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	freqs := make(map[int]int)
	loops := 10_000
	n := 4

	for _, item := range pop {
		freqs[item] = 0
	}

	for range loops {
		sampled := Sample(slices.Values(pop), n)
		assert.Len(t, sampled, n)
		for _, item := range sampled {
			freqs[item]++
		}
	}

	minCount := loops + 1
	maxCount := 0
	for _, count := range freqs {
		if count < minCount {
			minCount = count
		}
		if count > maxCount {
			maxCount = count
		}
	}

	maxMinRatio := float64(maxCount) / float64(minCount)
	assert.Less(t, maxMinRatio, 1.25)

	// Empty population
	empty := []int{}
	sampledEmpty := Sample(slices.Values(empty), 0)
	assert.Empty(t, sampledEmpty)

	// Sample size larger than population
	sampledLarge := Sample(slices.Values(pop), len(pop)+5)
	assert.Len(t, sampledLarge, len(pop))
}
