package radix

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/ywang2728/sampan/ds/linkedhashmap"
	"regexp"
	"testing"
)

func TestNewKeySeparator(t *testing.T) {
	tcs := []struct {
		bgn string
		end string
	}{
		{bgn: "(", end: ")"},
		{bgn: ":", end: "/"},
		{bgn: "{", end: "}"},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		assert.NotNil(t, ks)
		assert.Equal(t, ks.bs, tc.bgn)
		assert.Equal(t, ks.es, tc.end)
	}
}

func TestKeySeparatorIsOpenedAndIsClosed(t *testing.T) {
	tcs := []struct {
		bgn      string
		end      string
		char     string
		isOpened bool
		isClosed bool
	}{
		{bgn: "(", end: ")", char: "(", isOpened: true, isClosed: false},
		{bgn: ":", end: "/", char: "(", isOpened: false, isClosed: true},
		{bgn: "{", end: "}", char: "(", isOpened: false, isClosed: true},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		assert.False(t, ks.isOpened())
		assert.True(t, ks.isClosed())
		ks.openWith(tc.char)
		assert.Equal(t, tc.isOpened, ks.isOpened())
		assert.Equal(t, tc.isClosed, ks.isClosed())
	}
}

func TestKeySeparatorOpenWithAndCloseWith(t *testing.T) {
	tcs := []struct {
		bgn         string
		end         string
		bgnChars    []string
		endChars    []string
		closedTimes int
		isOpened    bool
		isClosed    bool
	}{
		{bgn: "(", end: ")", bgnChars: []string{"(", "("}, endChars: []string{")", ")"}, closedTimes: 0, isOpened: true, isClosed: true},
		{bgn: ":", end: "/", bgnChars: []string{":"}, endChars: []string{"/"}, closedTimes: 0, isOpened: true, isClosed: true},
		{bgn: "{", end: "}", bgnChars: []string{"{", "{"}, endChars: []string{"}", ")"}, closedTimes: 1, isOpened: true, isClosed: false},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		var times int
		var status bool
		for _, c := range tc.bgnChars {
			times, status = ks.openWith(c)
		}
		assert.Equal(t, len(tc.bgnChars), times)
		assert.Equal(t, tc.isOpened, status)
		for _, c := range tc.endChars {
			times, status = ks.closeWith(c)
		}
		assert.Equal(t, tc.closedTimes, times)
		assert.Equal(t, tc.isClosed, status)
	}
}

func TestKeySeparatorOpenAndClose(t *testing.T) {
	tcs := []struct {
		bgn           string
		end           string
		bgnChars      []string
		endChars      []string
		closedTimes   int
		isOpened      bool
		isClosed      bool
		isForceClosed bool
	}{
		{bgn: "{", end: "}", bgnChars: []string{"{", "{"}, endChars: []string{"}", ")"}, closedTimes: 0, isOpened: true, isClosed: false, isForceClosed: true},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		var times int
		var status bool
		for range tc.bgnChars {
			times, status = ks.open()
		}
		assert.Equal(t, len(tc.bgnChars), times)
		assert.Equal(t, tc.isOpened, status)
		for _, c := range tc.endChars {
			times, status = ks.closeWith(c)
		}
		assert.Equal(t, tc.isClosed, status)
		times, status = ks.close()
		assert.Equal(t, tc.closedTimes, times)
		assert.Equal(t, tc.isForceClosed, status)
		assert.Equal(t, tc.isForceClosed, ks.isClosed())
	}
}

func TestKeyIter(t *testing.T) {
	tcs := []struct {
		keys []Key[string]
	}{
		{
			keys: []Key[string]{},
		},
		{
			keys: []Key[string]{&staticKey{"abc"}},
		},
		{
			keys: []Key[string]{&staticKey{"abc"}, &wildcardStarKey{"*", "abc", "133", map[string]string{"*": ""}}},
		},
		{
			keys: []Key[string]{&staticKey{"abc"}, &wildcardStarKey{"*", "abc", "133", map[string]string{"*": ""}}, &regexKey{value: []string{"a", "{abc}", "123"}, patterns: linkedhashmap.New[string, *regexp.Regexp](), params: map[string]string{}}},
		},
	}
	for _, tc := range tcs {
		ki := newKeyIter(tc.keys...)
		assert.NotNil(t, ki)
		for _, k := range tc.keys {
			assert.True(t, ki.hasNext())
			assert.Equal(t, k, ki.Next())
		}
		assert.False(t, ki.hasNext())
	}
}

func TestBuildKeyIterFunc(t *testing.T) {
	tcs := []struct {
		key  string
		keys []Key[string]
	}{
		//{key: "", keys: []Key[string]{}},
		//{key: " ", keys: []Key[string]{}},
		//{key: "/", keys: []Key[string]{&staticKey{"/"}}},
		//{key: "abc", keys: []Key[string]{&staticKey{"abc"}}},
		//{key: "/abc", keys: []Key[string]{&staticKey{"/abc"}}},
		//{key: "abc/", keys: []Key[string]{&staticKey{"abc/"}}},
		//{key: "/123/abc", keys: []Key[string]{&staticKey{"/123/abc"}}},
		//{key: "/123/abc/", keys: []Key[string]{&staticKey{"/123/abc/"}}},
		//{key: "123/abc/", keys: []Key[string]{&staticKey{"123/abc/"}}},
		//{key: "*", keys: []Key[string]{&wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/*", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: "/123/*", keys: []Key[string]{&staticKey{"/123/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
	}

	for _, tc := range tcs {
		ki := buildKeyIterFunc(tc.key)
		assert.NotNil(t, ki)
		for _, k := range tc.keys {
			assert.True(t, ki.hasNext())
			a := ki.Next()
			fmt.Printf("result: %+v\n", a)
			assert.Equal(t, k, a)
		}
		assert.False(t, ki.hasNext())
	}
}

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
