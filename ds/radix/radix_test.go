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

func compileRePattern(raws ...string) (m *linkedhashmap.Map[string, *regexp.Regexp]) {
	m = linkedhashmap.New[string, *regexp.Regexp]()
	for _, raw := range raws {
		m.Put(raw, regexp.MustCompile(raw))
	}
	return m
}

func TestBuildKeyIterFunc(t *testing.T) {
	tcs := []struct {
		key  string
		keys []Key[string]
	}{
		{key: "", keys: []Key[string]{}},
		{key: " ", keys: []Key[string]{}},
		{key: "/", keys: []Key[string]{&staticKey{"/"}}},
		{key: "abc", keys: []Key[string]{&staticKey{"abc"}}},
		{key: "/abc", keys: []Key[string]{&staticKey{"/abc"}}},
		{key: "abc/", keys: []Key[string]{&staticKey{"abc/"}}},
		{key: "/123/abc", keys: []Key[string]{&staticKey{"/123/abc"}}},
		{key: "/123/abc/", keys: []Key[string]{&staticKey{"/123/abc/"}}},
		{key: "123/abc/", keys: []Key[string]{&staticKey{"123/abc/"}}},
		{key: "*", keys: []Key[string]{&wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "abc*", keys: []Key[string]{&wildcardStarKey{value: "*", prefix: "abc", params: map[string]string{"*": ""}}}},
		{key: "abc*123", keys: []Key[string]{&wildcardStarKey{value: "*", prefix: "abc", suffix: "123", params: map[string]string{"*": ""}}}},
		{key: "/*", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/abc*", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", prefix: "abc", params: map[string]string{"*": ""}}}},
		{key: "/abc*123", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", prefix: "abc", suffix: "123", params: map[string]string{"*": ""}}}},
		{key: "*/", keys: []Key[string]{&wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "abc*/", keys: []Key[string]{&wildcardStarKey{value: "*", prefix: "abc", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "abc*123/", keys: []Key[string]{&wildcardStarKey{value: "*", prefix: "abc", suffix: "123", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/*/", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/abc*/", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", prefix: "abc", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/abc*123/", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", prefix: "abc", suffix: "123", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/123/*", keys: []Key[string]{&staticKey{"/123/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "*/123", keys: []Key[string]{&wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/123"}}},
		{key: "*/123/", keys: []Key[string]{&wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: "/*/123/", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: "/abc/*/123/", keys: []Key[string]{&staticKey{"/abc/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: "/abc/def*/123/", keys: []Key[string]{&staticKey{"/abc/"}, &wildcardStarKey{value: "*", prefix: "def", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: "/abc/*hij/123/", keys: []Key[string]{&staticKey{"/abc/"}, &wildcardStarKey{value: "*", suffix: "hij", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: "/abc/def*hij/123/", keys: []Key[string]{&staticKey{"/abc/"}, &wildcardStarKey{value: "*", prefix: "def", suffix: "hij", params: map[string]string{"*": ""}}, &staticKey{"/123/"}}},
		{key: ":abc", keys: []Key[string]{&wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}}},
		{key: "/:abc", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}}},
		{key: ":abc/", keys: []Key[string]{&wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		{key: "/:abc/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		{key: "/:abc/123/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/123/"}}},
		{key: "/123/:abc/", keys: []Key[string]{&staticKey{"/123/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		{key: "/123/:abc/789/", keys: []Key[string]{&staticKey{"/123/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/789/"}}},
		{key: "{abc}", keys: []Key[string]{&regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		{key: "/{abc}", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		{key: "/{abc}/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		{key: "123{abc}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		{key: "123{abc}789", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		{key: "/123{abc}789", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		{key: "123{abc}789/", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		{key: "/123{abc}789/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		{key: "/123{abc}789{def}/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}, &staticKey{"/"}}},
		{key: "toto/123{abc}789{def}", keys: []Key[string]{&staticKey{"toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}}},
		{key: "/toto/123{abc}789{def}", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}}},
		{key: "/toto/123{abc}789{def}/hoho/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}, &staticKey{"/hoho/"}}},
		{key: "{abc}/*", keys: []Key[string]{&regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: ":xyz/{abc}/*", keys: []Key[string]{&wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/:xyz/{abc}/*", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/:xyz/{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/123{abc}/*", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/123{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/toto/123{abc}/*/hoho", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/hoho"}}},
		{key: "/toto/123{abc}789/:xyz/hoho/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hoho/"}}},
		{key: "/toto/123{abc}789/:xyz/hoho/pre*", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hoho/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}}},
		{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789/:xyz/hoho/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hoho/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789-{\\w+}/:xyz/hoho/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789-", "{\\w+}"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hoho/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "123{abc}789{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		{key: "123{abc}{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		{key: "123{abc}789{(?P<date>[a-z][0-9]?)}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}}},
		{key: "123{abc}789{(?P<date>[a-z][0-9]?)}-{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}", "-", "{\\w+}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}}},
		{key: "hello-{(?P<abc>[a-z]+)}!=bonjour-{(?P<def>\\d\\w*)}-world/", keys: []Key[string]{&regexKey{value: []string{"hello-", "{(?P<abc>[a-z]+)}", "!=bonjour-", "{(?P<def>\\d\\w*)}", "-world"}, patterns: compileRePattern("{(?P<abc>[a-z]+)}", "{(?P<def>\\d\\w*)}"), params: map[string]string{"abc": "", "def": ""}}, &staticKey{"/"}}},
	}

	for _, tc := range tcs {
		ki := buildKeyIterFunc(tc.key)
		assert.NotNil(t, ki)
		for _, k := range tc.keys {
			assert.True(t, ki.hasNext())
			assert.Equal(t, k, ki.Next())
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
