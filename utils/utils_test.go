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
		assert.Equal(t, tc.index, IndexNth(tc.key, tc.char, tc.occur))
	}
}

func TestInfix2Suffix(t *testing.T) {
	tcs := []struct {
		infix  string
		suffix []string
	}{
		{"2+3*(7-4)+8/4", []string{"2", "3", "7", "4", "-", "*", "+", "8", "4", "/", "+"}},
		{"((2+3)*4-(8+2))/5", []string{"2", "3", "+", "4", "*", "8", "2", "+", "-", "5", "/"}},
		{"1314+25.5*12", []string{"1314", "25.5", "12", "*", "+"}},
		{"-2*(+3)", []string{"-2", "3", "*"}},
		{"-2*(+3)-10", []string{"-2", "3", "*", "10", "-"}},
		{"123", []string{"123"}},
		{"-123", []string{"-123"}},
	}
	for _, tc := range tcs {
		assert.Equal(t, tc.suffix, Infix2Suffix(&tc.infix))
	}

}
