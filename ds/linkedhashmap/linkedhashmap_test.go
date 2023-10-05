package linkedhashmap

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapNew(t *testing.T) {
	m := New[string, string]()
	assert.NotNil(t, m)
	assert.NotNil(t, m.index)
	assert.NotNil(t, m.dict)
}

type kv struct {
	k string
	v string
}

func TestMapPut(t *testing.T) {
	tcs := []struct {
		kvList     []kv
		kvExpected []kv
	}{
		{
			kvList:     []kv{{k: "a", v: "1"}},
			kvExpected: []kv{{k: "a", v: "1"}},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "b", v: "2"}, {k: "e", v: "5"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvExpected), m.len())
		mi := m.Iter()
		for _, kv := range tc.kvExpected {
			if mi.HasNext() {
				mk, mv := mi.Next()
				assert.Equal(t, kv.k, mk)
				assert.Equal(t, kv.v, mv)
			}
		}

	}
}

func TestMapGet(t *testing.T) {
	tcs := []struct {
		kvList     []kv
		kvExpected []kv
	}{
		{
			kvList:     []kv{{k: "a", v: "1"}},
			kvExpected: []kv{{k: "a", v: "1"}},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "b", v: "2"}, {k: "e", v: "5"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvExpected), m.len())
		for _, kv := range tc.kvExpected {
			v, _ := m.Get(kv.k)
			assert.Equal(t, kv.v, v)
		}
	}
}

func TestMapRemove(t *testing.T) {
	tcs := []struct {
		kvList     []kv
		kvRemove   []kv
		kvExpected []kv
	}{
		{
			kvList:     []kv{{k: "a", v: "1"}},
			kvRemove:   []kv{{k: "a", v: "1"}},
			kvExpected: []kv{},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			kvRemove:   []kv{{k: "b", v: "2"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "c", v: "3"}},
		},
		{
			kvList:     []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "b", v: "2"}, {k: "e", v: "5"}},
			kvRemove:   []kv{{k: "b", v: "2"}, {k: "c", v: "3"}},
			kvExpected: []kv{{k: "a", v: "1"}, {k: "e", v: "5"}},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		for _, kv := range tc.kvRemove {
			m.Remove(kv.k)
		}
		assert.Equal(t, len(tc.kvExpected), m.len())
		for _, kv := range tc.kvExpected {
			v, _ := m.Get(kv.k)
			assert.Equal(t, kv.v, v)
		}
	}
}

func TestMapClear(t *testing.T) {
	tcs := []struct {
		kvList []kv
	}{
		{
			kvList: []kv{{k: "a", v: "1"}},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvList), m.len())
		m.clear()
		assert.Equal(t, 0, m.len())
	}
}

func TestMapKeys(t *testing.T) {
	tcs := []struct {
		kvList []kv
		kvKeys []string
	}{
		{
			kvList: []kv{{k: "a", v: "1"}},
			kvKeys: []string{"a"},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			kvKeys: []string{"a", "b", "c"},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
			kvKeys: []string{"a", "b", "c", "e"},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvList), m.len())
		assert.EqualValues(t, tc.kvKeys, m.Keys())
	}
}

func TestMapValues(t *testing.T) {
	tcs := []struct {
		kvList   []kv
		kvValues []string
	}{
		{
			kvList:   []kv{{k: "a", v: "1"}},
			kvValues: []string{"1"},
		},
		{
			kvList:   []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			kvValues: []string{"1", "2", "3"},
		},
		{
			kvList:   []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
			kvValues: []string{"1", "2", "3", "5"},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvList), m.len())
		assert.EqualValues(t, tc.kvValues, m.Values())
	}
}

func TestMapString(t *testing.T) {
	tcs := []struct {
		kvList []kv
		output string
	}{
		{
			kvList: []kv{{k: "a", v: "1"}},
			output: "LinkedHashMap[a:1]",
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
			output: "LinkedHashMap[a:1 b:2 c:3]",
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
			output: "LinkedHashMap[a:1 b:2 c:3 e:5]",
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		assert.Equal(t, len(tc.kvList), m.len())
		assert.Equal(t, tc.output, m.String())
		assert.Equal(t, tc.output, fmt.Sprint(m))
	}
}

func TestMapIter(t *testing.T) {
	tcs := []struct {
		kvList []kv
	}{
		{
			kvList: []kv{},
		},
		{
			kvList: []kv{{k: "a", v: "1"}},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}},
		},
		{
			kvList: []kv{{k: "a", v: "1"}, {k: "b", v: "2"}, {k: "c", v: "3"}, {k: "e", v: "5"}},
		},
	}
	for _, tc := range tcs {
		m := New[string, string]()
		for _, kv := range tc.kvList {
			m.Put(kv.k, kv.v)
		}
		mi := m.Iter()
		if len(tc.kvList) == 0 {
			assert.False(t, mi.HasPrev())
			assert.False(t, mi.HasNext())
		} else if len(tc.kvList) == 1 {
			assert.False(t, mi.HasPrev())
			assert.True(t, mi.HasNext())
			k, v := mi.Next()
			assert.Equal(t, tc.kvList[0].k, k)
			assert.Equal(t, tc.kvList[0].v, v)
			assert.False(t, mi.HasPrev())
			assert.False(t, mi.HasNext())
		} else {
			for i := 0; i < len(tc.kvList); i++ {
				k, v := mi.Next()
				if i == 0 {
					assert.False(t, mi.HasPrev())
					assert.True(t, mi.HasNext())
				} else if i == len(tc.kvList)-1 {
					assert.True(t, mi.HasPrev())
					assert.False(t, mi.HasNext())
				} else {
					assert.True(t, mi.HasPrev())
					assert.True(t, mi.HasNext())
				}
				assert.Equal(t, tc.kvList[i].k, k)
				assert.Equal(t, tc.kvList[i].v, v)
			}
		}
	}
}
