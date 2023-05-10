package log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseLevel(t *testing.T) {
	tcs := []struct {
		l     string
		level int
	}{
		{"off", 0},
		{"fatal", 1},
		{"FataL", 1},
		{"FATAL", 1},
		{"error", 2},
		{"warn", 3},
		{"INFO", 4},
		{"DEBUG", 5},
		{"TRACE", 6},
		{"unknown", 0},
	}
	for _, tc := range tcs {
		ll := parseLevel(tc.l)
		assert.Equal(t, tc.level, ll)
	}
}
