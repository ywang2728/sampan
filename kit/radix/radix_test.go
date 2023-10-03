package radix

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type (
	wildcardParser struct {
		raw string
	}
)

func (wp *wildcardParser) compare(k1, k2 string) (bool, string) {
	return k1 == k2, "haha"
}

var wcParser = wildcardParser{}

func TestNewRadix(t *testing.T) {
	r := New[string, func(int)]
	assert.NotNil(t, r)
	assert.Nil(t, r)
}
