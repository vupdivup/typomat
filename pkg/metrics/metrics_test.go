package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWPM(t *testing.T) {
	cases := []struct {
		input    string
		duration time.Duration
		want     float64
	}{
		{"", 0 * time.Minute, 0.0},
		{"lorem ipsum", 1 * time.Minute, 2},
		{"lorem ipsum dolor sit am", 10 * time.Second, 24},
		{"  lorem i      ", 1 * time.Second, 72},
	}

	for _, c := range cases {
		got := WPM(c.input, c.duration)
		assert.Equal(t, c.want, got)
	}
}

func TestAccuracy(t *testing.T) {
	cases := []struct {
		prompt string
		input  string
		want   float64
	}{
		// empty prompt & input
		{"", "", 100.0},
		// completely wrong
		{"abcde", "fghij", 0.0},
		// perfect match
		{"hello", "hello", 100.0},
		// single mismatch (4/5 correct)
		{"hello", "hxllo", 80.0},
		// three mismatches (7/10 correct)
		{"abcdefghij", "abc123ghij", 70.0},
		// input shorter (first 5 all correct)
		{"abcdefghij", "abcde", 100.0},
		// input longer (only first 5 measured)
		{"abcde", "abcdeZZZZ", 100.0},
		// last char mismatch
		{"abcde", "abcdf", 80.0},
		// unicode with one mismatch (3/4 correct)
		{"你好世界", "你好世X", 75.0},
	}

	for _, c := range cases {
		got := Accuracy(c.prompt, c.input)
		assert.Equal(t, c.want, got)
	}
}
