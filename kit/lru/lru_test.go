package lru

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type (
	testValue struct {
		data   int
		params map[string]string
	}
)

func TestNew(t *testing.T) {
	cache := New[string, any](1)
	assert.NotNil(t, cache)
	assert.NotNil(t, cache.list)
	assert.NotNil(t, cache.keys)
}

func TestPutAndLen(t *testing.T) {
	cache := New[string, int](3)
	cache.put("hello", 1)
	cache.put("world", 2)
	assert.Equal(t, 2, cache.len())
}

func TestPutAndGet(t *testing.T) {
	cache := New[string, int](3)
	values := [3]string{"a", "b", "c"}
	for i, v := range values {
		cache.put(v, i)
	}
	assert.Equal(t, 3, cache.len())
	cache.put("d", 4)
	assert.Equal(t, 3, cache.len())
	value, ok := cache.get("a")
	assert.Equal(t, 0, value)
	assert.False(t, ok)
	cache.get("c")
	cache.put("e", 5)
	assert.Equal(t, 3, cache.len())
	value, ok = cache.get("b")
	assert.Equal(t, 0, value)
	assert.False(t, ok)
	cache.get("d")
	cache.get("e")
	cache.put("f", 6)
	value, ok = cache.get("e")
	assert.Equal(t, 5, value)
	assert.True(t, ok)
	value, ok = cache.get("c")
	assert.Equal(t, 0, value)
	assert.False(t, ok)
}

func TestDelete(t *testing.T) {
	cache := New[string, int](3)
	values := [3]string{"a", "b", "c"}
	for i, v := range values {
		cache.put(v, i)
	}
	assert.Equal(t, 3, cache.len())
	cache.delete("c")
	assert.Equal(t, 2, cache.len())
	value, ok := cache.get("c")
	assert.Equal(t, 0, value)
	assert.False(t, ok)
}
