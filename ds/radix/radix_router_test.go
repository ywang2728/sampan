package radix

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func buildStaticKeyIter(ss ...string) (ki KeyIterator[string]) {
	var keys []Key[string]
	for _, s := range ss {
		keys = append(keys, &staticKey{s})
	}
	return &keyIter{-1, keys}
}

func TestStaticKeyString(t *testing.T) {
	tcs := []struct {
		value  string
		output string
	}{
		{value: "/", output: "/"},
		{value: "/abc", output: "/abc"},
		{value: "/123", output: "/123"},
		{value: "/123/abc", output: "/123/abc"},
		{value: "123/", output: "123/"},
		{value: "123", output: "123"},
		{value: "123abc", output: "123abc"},
		{value: "123/abc", output: "123/abc"},
	}
	for _, tc := range tcs {
		var sk Key[string] = &staticKey{value: tc.value}
		assert.Equal(t, tc.output, sk.String())
		assert.Equal(t, tc.output, fmt.Sprint(sk))
	}
}

func TestStaticKeyMatchIterator(t *testing.T) {
	tcs := []struct {
		value    string
		path     KeyIterator[string]
		common   KeyIterator[string]
		tailKey  KeyIterator[string]
		tailPath KeyIterator[string]
	}{
		{value: "/", path: buildStaticKeyIter("/abc"), common: buildStaticKeyIter("/"), tailKey: nil, tailPath: buildStaticKeyIter("abc")},
		{value: "/", path: buildStaticKeyIter("/", "abc"), common: buildStaticKeyIter("/"), tailKey: nil, tailPath: buildStaticKeyIter("abc")},
		{value: "/", path: buildStaticKeyIter("/123", "abc"), common: buildStaticKeyIter("/"), tailKey: nil, tailPath: buildStaticKeyIter("123abc")},
		{value: "/123", path: buildStaticKeyIter("/pic", "nic"), common: buildStaticKeyIter("/"), tailKey: buildStaticKeyIter("123"), tailPath: buildStaticKeyIter("picnic")},
		{value: "/pic", path: buildStaticKeyIter("/pic", "nic"), common: buildStaticKeyIter("/pic"), tailKey: nil, tailPath: buildStaticKeyIter("nic")},
		{value: "/pic", path: buildStaticKeyIter("/picture", "/nic"), common: buildStaticKeyIter("/pic"), tailKey: nil, tailPath: buildStaticKeyIter("ture/nic")},
		{value: "/abc", path: buildStaticKeyIter("/"), common: buildStaticKeyIter("/"), tailKey: buildStaticKeyIter("abc"), tailPath: nil},
		{value: "/picnic", path: buildStaticKeyIter("/abc", "/hello", "world"), common: buildStaticKeyIter("/"), tailKey: buildStaticKeyIter("picnic"), tailPath: buildStaticKeyIter("abc/hello", "world")},
		{value: "/123", path: buildStaticKeyIter("/123abc"), common: buildStaticKeyIter("/123"), tailKey: nil, tailPath: buildStaticKeyIter("abc")},
		{value: "/123/", path: buildStaticKeyIter("/123/abc"), common: buildStaticKeyIter("/123/"), tailKey: nil, tailPath: buildStaticKeyIter("abc")},
		{value: "123/", path: buildStaticKeyIter("/123/abc"), common: nil, tailKey: buildStaticKeyIter("123/"), tailPath: buildStaticKeyIter("/123/abc")},
		{value: "123/", path: buildStaticKeyIter("/123", "/abc"), common: nil, tailKey: buildStaticKeyIter("123/"), tailPath: buildStaticKeyIter("/123", "/abc")},
		{value: "123", path: buildStaticKeyIter("/123/abc"), common: nil, tailKey: buildStaticKeyIter("123"), tailPath: buildStaticKeyIter("/123/abc")},
		{value: "123abc", path: buildStaticKeyIter("123/abc"), common: buildStaticKeyIter("123"), tailKey: buildStaticKeyIter("abc"), tailPath: buildStaticKeyIter("/abc")},
		{value: "123abc", path: buildStaticKeyIter("123/abc", "def"), common: buildStaticKeyIter("123"), tailKey: buildStaticKeyIter("abc"), tailPath: buildStaticKeyIter("/abcdef")},
		{value: "123/abc", path: buildStaticKeyIter("123/"), common: buildStaticKeyIter("123/"), tailKey: buildStaticKeyIter("abc"), tailPath: nil},
	}
	for _, tc := range tcs {
		var sk Key[string] = &staticKey{value: tc.value}
		assert.Equal(t, tc.value, sk.(*staticKey).value)
		c, tk, tp := sk.MatchIterator(tc.path)
		if tc.common == nil {
			assert.Nil(t, c)
		} else {
			assert.True(t, c.HasNext())
			assert.Equal(t, tc.common.Next().(*staticKey).value, c.Next().(*staticKey).value)
		}
		if tc.tailKey == nil {
			assert.Nil(t, tk)
		} else {
			assert.True(t, tk.HasNext())
			assert.Equal(t, tc.tailKey.Next().(*staticKey).value, tk.Next().(*staticKey).value)
		}
		if tc.tailPath == nil {
			assert.Nil(t, tp)
		} else {
			assert.True(t, tp.HasNext())
			assert.Equal(t, tc.tailPath.Next().(*staticKey).value, tp.Next().(*staticKey).value)
		}

	}
}

//func TestStaticKeyMatch(t *testing.T) {
//	tcs := []struct {
//		value   string
//		path    string
//		tail    string
//		params  map[string]string
//		matched bool
//	}{
//		{value: "/", path: "/abc", common: buildKeyIter("/"), tailKey: nil, tailPath: buildKeyIter("abc")},
//		{value: "/abc", path: "/", common: buildKeyIter("/"), tailKey: buildKeyIter("abc"), tailPath: nil},
//		{value: "/abc", path: "/ab", common: buildKeyIter("/ab"), tailKey: buildKeyIter("c"), tailPath: nil},
//		{value: "/123", path: "/123abc", common: buildKeyIter("/123"), tailKey: nil, tailPath: buildKeyIter("abc")},
//		{value: "/123/", path: "/123/abc", common: buildKeyIter("/123/"), tailKey: nil, tailPath: buildKeyIter("abc")},
//		{value: "123/", path: "/123/abc", common: nil, tailKey: buildKeyIter("123/"), tailPath: buildKeyIter("/123/abc")},
//		{value: "123", path: "/123/abc", common: nil, tailKey: buildKeyIter("123"), tailPath: buildKeyIter("/123/abc")},
//		{value: "123abc", path: "123/abc", common: buildKeyIter("123"), tailKey: buildKeyIter("abc"), tailPath: buildKeyIter("/abc")},
//		{value: "123/abc", path: "123/", common: buildKeyIter("123/"), tailKey: buildKeyIter("abc"), tailPath: nil},
//	}
//	for _, tc := range tcs {
//		var sk Key[string] = &staticKey{value: tc.value}
//		assert.Equal(t, tc.value, sk.Value())
//		c, tk, tp, _ := sk.Match(&staticKey{value: tc.path})
//		if tc.common == nil {
//			assert.Nil(t, c)
//		} else {
//			assert.True(t, c.HasNext())
//			assert.Equal(t, tc.common.Next().Value(), c.Next().Value())
//		}
//		if tc.tailKey == nil {
//			assert.Nil(t, tk)
//		} else {
//			assert.True(t, tk.HasNext())
//			assert.Equal(t, tc.tailKey.Next().Value(), tk.Next().Value())
//		}
//		if tc.tailPath == nil {
//			assert.Nil(t, tp)
//		} else {
//			assert.True(t, tp.HasNext())
//			assert.Equal(t, tc.tailPath.Next().Value(), tp.Next().Value())
//		}
//
//	}
//}
