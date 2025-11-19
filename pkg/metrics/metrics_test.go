package metrics

import (
	"testing"
	"time"
)

func TestWPM(t *testing.T) {
	cases := []struct {
		input    string
		duration time.Duration
		want     float64
	}{
		{"lorem ipsum", 1 * time.Minute, 2},
		{"lorem ipsum dolor sit am", 10 * time.Second, 24},
		{"  lorem i      ", 1 * time.Second, 72},
	}

	for _, c := range cases {
		got := WPM(c.input, c.duration)
		if got != c.want {
			t.Errorf(
				"WPM(%q, %v) = %v; want %v", c.input, c.duration, got, c.want,
			)
		}
	}
}

func TestAccuracy(t *testing.T) {
	cases := []struct {
		prompt string
		input  string
		want   float64
	}{
		{"", "", 100.0},
		{"abcde", "fghij", 0.0},
		{"abcde", "abcde", 100.0},
		{"abcde", "abcdf", 80.0},
		{"   abcde  ", "  abfde   ", 40.0},
		{"abcd", "abc", 75.0},
	}

	for _, c := range cases {
		got := Accuracy(c.prompt, c.input)
		if got != c.want {
			t.Errorf(
				"Accuracy(%q, %q) = %v; want %v",
				c.prompt, c.input, got, c.want,
			)
		}
	}
}
