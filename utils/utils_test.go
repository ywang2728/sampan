package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIndexNth(t *testing.T) {
	tcs := []struct {
		key   string
		char  uint8
		occur int
		index int
	}{
		{"abcde", 's', 1, -1},
		{"abcde", 'b', 1, 1},
		{"abcbe", 'b', 2, 3},
		{"abcbb", 'b', 2, 3},
		{"abcbb", 'b', 3, 4},
		{"abcbb", 'b', 5, -1},
		{"abcbb", 'c', 2, -1},
		{"abcbb", 'c', 1, 2},
	}
	for _, tc := range tcs {
		assert.Equal(t, tc.index, indexNth(tc.key, tc.char, tc.occur))
	}
}
