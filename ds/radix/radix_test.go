package radix

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringKey(t *testing.T) {
	tcs := []struct {
		value    string
		path     string
		common   KeyIterator[string]
		tailKey  KeyIterator[string]
		tailPath KeyIterator[string]
	}{
		{value: "/", path: "/abc", common: buildKeyIterFunc("/"), tailKey: nil, tailPath: buildKeyIterFunc("abc")},
		{value: "/abc", path: "/", common: buildKeyIterFunc("/"), tailKey: buildKeyIterFunc("abc"), tailPath: nil},
		{value: "/abc", path: "/ab", common: buildKeyIterFunc("/ab"), tailKey: buildKeyIterFunc("c"), tailPath: nil},
		{value: "/123", path: "/123abc", common: buildKeyIterFunc("/123"), tailKey: nil, tailPath: buildKeyIterFunc("abc")},
		{value: "/123/", path: "/123/abc", common: buildKeyIterFunc("/123/"), tailKey: nil, tailPath: buildKeyIterFunc("abc")},
		{value: "123/", path: "/123/abc", common: nil, tailKey: buildKeyIterFunc("123/"), tailPath: buildKeyIterFunc("/123/abc")},
		{value: "123", path: "/123/abc", common: nil, tailKey: buildKeyIterFunc("123"), tailPath: buildKeyIterFunc("/123/abc")},
		{value: "123abc", path: "123/abc", common: buildKeyIterFunc("123"), tailKey: buildKeyIterFunc("abc"), tailPath: buildKeyIterFunc("/abc")},
		{value: "123/abc", path: "123/", common: buildKeyIterFunc("123/"), tailKey: buildKeyIterFunc("abc"), tailPath: nil},
	}
	for _, tc := range tcs {
		var sk Key[string] = &staticKey{value: tc.value}
		assert.Equal(t, tc.value, fmt.Sprint(sk))
		c, tk, tp, _ := sk.Match(tc.path)
		if tc.common == nil {
			assert.Nil(t, c)
		} else {
			assert.Equal(t, tc.common.Next().Value(), c.Next().Value())
		}
		if tc.tailKey == nil {
			assert.Nil(t, tk)
		} else {
			assert.Equal(t, tc.tailKey.Next().Value(), tk.Next().Value())
		}
		if tc.tailPath == nil {
			assert.Nil(t, tp)
		} else {
			assert.Equal(t, tc.tailPath.Next().Value(), tp.Next().Value())
		}

	}
}

func TestNewRadix(t *testing.T) {
	r := New[string, func()](buildKeyIterFunc)
	assert.Empty(t, r.root)
	assert.Empty(t, r.size)
	assert.NotNil(t, r.buildKeyIter)
}

func TestNewNode(t *testing.T) {
	r := New[string, func()](buildKeyIterFunc)
	ki := r.buildKeyIter("aaa")
	n := r.newNode(ki.Next())
	assert.NotNil(t, n.k)
	assert.Nil(t, n.v)
	assert.Empty(t, n.nodes)
}

func TestPutRecWithStringKeySingleNode(t *testing.T) {
	r := New[string, func()](buildKeyIterFunc)
	assert.True(t, r.put("/aaa", hello))
	assert.NotNil(t, r.root)
	assert.Equal(t, 1, r.Len())
	assert.Equal(t, "/aaa", fmt.Sprint(r.root.k))
}
