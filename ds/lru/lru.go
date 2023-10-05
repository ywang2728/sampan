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
		dict  map[K]*list.Element
		mutex sync.RWMutex
	}
)

func New[K comparable, V any](cap int) *Cache[K, V] {
	return &Cache[K, V]{
		cap:  cap,
		list: list.New(),
		dict: make(map[K]*list.Element, cap),
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
	clear(c.dict)
}

func (c *Cache[K, V]) put(key K, value V) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.dict[key]; ok {
		c.list.MoveToFront(e)
	} else {
		if c.list.Len() == c.cap {
			delete(c.dict, c.list.Remove(c.list.Back()).(*element[K, V]).key)
		}
		c.dict[key] = c.list.PushFront(&element[K, V]{key: key, value: value})
	}
}

func (c *Cache[K, V]) get(key K) (value V, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, exist := c.dict[key]; exist {
		c.list.MoveToFront(e)
		value = e.Value.(*element[K, V]).value
		return value, true
	}
	return value, false
}

func (c *Cache[K, V]) delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if e, ok := c.dict[key]; ok {
		delete(c.dict, c.list.Remove(e).(*element[K, V]).key)
	}
}
