package linkedhashmap

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type (
	Map[K comparable, V any] struct {
		index *list.List
		dict  map[K]V
		sync.RWMutex
	}
	MapIterator[K comparable, V any] struct {
		cursor  int
		element *list.Element
		dict    *Map[K, V]
	}
)

func (mi *MapIterator[K, V]) HasPrev() bool {
	return mi.cursor-1 < mi.dict.len() && mi.cursor-1 >= 0
}

func (mi *MapIterator[K, V]) Prev() (k K, v V) {
	if mi.HasPrev() {
		mi.element = mi.element.Prev()
		mi.cursor--
		k = mi.element.Value.(K)
		v, _ = mi.dict.Get(k)
	}
	return
}

func (mi *MapIterator[K, V]) HasNext() bool {
	return mi.cursor+1 < mi.dict.len() && mi.cursor+1 >= 0
}

func (mi *MapIterator[K, V]) Next() (k K, v V) {
	if mi.HasNext() {
		if mi.cursor == -1 {
			mi.element = mi.dict.index.Front()
		} else {
			mi.element = mi.element.Next()
		}
		mi.cursor++
		k = mi.element.Value.(K)
		v, _ = mi.dict.Get(k)
	}
	return
}

func New[K comparable, V any]() *Map[K, V] {
	return &Map[K, V]{
		index: list.New(),
		dict:  map[K]V{},
	}
}

func NewFromMap[K comparable, V any](m map[K]V) (lm *Map[K, V]) {
	lm = New[K, V]()
	for k, v := range m {
		lm.Put(k, v)
	}
	return
}

func (m *Map[K, V]) len() int {
	m.RLock()
	defer m.RUnlock()
	return m.index.Len()
}

func (m *Map[K, V]) clear() {
	m.Lock()
	defer m.Unlock()
	m.index.Init()
	clear(m.dict)
}

func (m *Map[K, V]) Put(key K, value V) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.dict[key]; !ok {
		m.index.PushBack(key)
	}
	m.dict[key] = value
}

func (m *Map[K, V]) Get(key K) (value V, ok bool) {
	m.Lock()
	defer m.Unlock()
	value, ok = m.dict[key]
	return
}

func (m *Map[K, V]) Remove(key K) {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.dict[key]; ok {
		for e := m.index.Front(); e != nil; e = e.Next() {
			k := e.Value.(K)
			if k == key {
				m.index.Remove(e)
				delete(m.dict, k)
				break
			}
		}
	}
}

func (m *Map[K, V]) Keys() (keys []K) {
	for e := m.index.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(K))
	}
	return
}

func (m *Map[K, V]) Values() (values []V) {
	for e := m.index.Front(); e != nil; e = e.Next() {
		values = append(values, m.dict[e.Value.(K)])
	}
	return
}

func (m *Map[K, V]) String() string {
	var sb strings.Builder
	sb.WriteString("LinkedHashMap[")
	for e := m.index.Front(); e != nil; e = e.Next() {
		sb.WriteString(fmt.Sprintf("%v:%v ", e.Value, m.dict[e.Value.(K)]))
	}
	return strings.TrimRight(sb.String(), " ") + "]"
}

func (m *Map[K, V]) Contains(k K) (ok bool) {
	_, ok = m.dict[k]
	return
}

func (m *Map[K, V]) Iter() MapIterator[K, V] {
	return MapIterator[K, V]{
		cursor: -1,
		dict:   m,
	}
}
