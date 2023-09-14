package lru

import (
	"container/list"
	"sync"
)

type (
	element[K comparable, V any] struct {
		key   K
		value V
	}

	Cache[K comparable, V any] struct {
		cap   int
		list  *list.List
		keys  map[K]*list.Element
		mutex sync.RWMutex
	}
)

func New[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:  cap,
		list: list.New(),
		keys: make(map[K]*list.Element, cap),
	}
}

func (c *Cache[K, V]) len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.list.Len()
}

func (c *Cache[K, V]) clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.list.Init()
	clear(c.keys)
}

func (c *Cache[K, V]) put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.keys[key]; ok {
		c.list.MoveToFront(e)
	} else {
		if c.list.Len() == c.cap {
			delete(c.keys, c.list.Remove(c.list.Back()).(*element[K, V]).key)
		}
		c.keys[key] = c.list.PushFront(&element[K, V]{key: key, value: value})
	}
}

func (c *Cache[K, V]) get(key K) (value V, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, exist := c.keys[key]; exist {
		c.list.MoveToFront(e)
		value = e.Value.(*element[K, V]).value
		return value, true
	}
	return value, false
}

func (c *Cache[K, V]) delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.keys[key]; ok {
		delete(c.keys, c.list.Remove(e).(*element[K, V]).key)
	}
}
