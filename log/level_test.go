package log

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLevel_NameAndIndex(t *testing.T) {
	tcs := []struct {
		l     Level
		name  string
		index int
	}{
		{OFF, "OFF", 0},
		{FATAL, "FATAL", 1},
		{ERROR, "ERROR", 2},
		{WARN, "WARN", 3},
		{INFO, "INFO", 4},
		{DEBUG, "DEBUG", 5},
		{TRACE, "TRACE", 6},
	}
	for _, tc := range tcs {
		assert.Equal(t, tc.name, tc.l.Name())
		assert.Equal(t, tc.index, tc.l.Index())
	}
}
