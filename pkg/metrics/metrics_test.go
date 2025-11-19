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
