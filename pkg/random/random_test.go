package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShuffle(t *testing.T) {
	original := []int{1, 2, 3, 4, 5}
	freqs := make(map[int]map[int]int)
	loops := 10_000

	for _, item := range original {
		freqs[item] = make(map[int]int)
		for i := range len(original) {
			freqs[item][i] = 0
		}
	}

	for range loops {
		shuffled := Shuffle(original)
		for i, item := range shuffled {
			freqs[item][i]++
		}
	}

	for _, positions := range freqs {
		minCount := loops + 1
		maxCount := 0
		for _, count := range positions {
			if count < minCount {
				minCount = count
			}
			if count > maxCount {
				maxCount = count
			}
		}

		maxMinRatio := float64(maxCount) / float64(minCount)
		assert.Less(t, maxMinRatio, 1.2)
	}

	// Empty slice
	empty := []int{}
	shuffledEmpty := Shuffle(empty)
	assert.Empty(t, shuffledEmpty)
}
