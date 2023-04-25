package sampan

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	assert.NotNil(t, s)
	assert.NotNil(t, s.rg)
}
