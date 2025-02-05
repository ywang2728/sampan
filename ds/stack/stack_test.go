package stack

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMutexStackNew(t *testing.T) {
	s := NewMutexStack[string]()
	assert.NotNil(t, s)
	assert.Nil(t, s.head)
	assert.Nil(t, s.tail)
}

func TestStackIsEmpty(t *testing.T) {
	s := NewMutexStack[string]()
	assert.True(t, s.IsEmpty())
	s.Push("haha")
	assert.False(t, s.IsEmpty())
}

func TestStackPushPeekAndPop(t *testing.T) {
	tcs := []struct {
		pushList []string
	}{
		{pushList: []string{}},
		{pushList: []string{"a", "b", "c"}},
		{pushList: []string{"abc", "bdd"}},
	}
	for _, tc := range tcs {
		s := NewMutexStack[string]()
		for _, elem := range tc.pushList {
			s.Push(elem)
		}
		reversed := make([]string, len(tc.pushList))
		copy(reversed, tc.pushList[:])
		for i, j := 0, len(reversed)-1; i < j; i, j = i+1, j-1 {
			reversed[i], reversed[j] = reversed[j], reversed[i]
		}
		for _, elem := range reversed {
			e, ok := s.Peek()
			assert.True(t, ok)
			assert.Equal(t, elem, e)
			e, ok = s.Pop()
			assert.True(t, ok)
			assert.Equal(t, elem, e)
		}
		assert.True(t, s.IsEmpty())
	}
}

func TestStackString(t *testing.T) {
	tcs := []struct {
		pushList []string
		output   string
	}{
		{pushList: []string{}, output: "MutexStack[]"},
		{pushList: []string{"a", "b", "c"}, output: "MutexStack[a b c]"},
		{pushList: []string{"abc", "bdd"}, output: "MutexStack[abc bdd]"},
	}
	for _, tc := range tcs {
		s := NewMutexStack[string]()
		for _, elem := range tc.pushList {
			s.Push(elem)
		}
		assert.Equal(t, tc.output, s.String())
	}
}
