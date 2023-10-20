package radix

import (
	"github.com/stretchr/testify/assert"
	"github.com/ywang2728/sampan/ds/linkedhashmap"
	"github.com/ywang2728/sampan/ds/stack"
	"regexp"
	"testing"
)

type (
	testNode struct {
		key   string
		value func()
	}
)

func depthFirst(r *Radix[string, func()]) (df []string) {
	s := stack.New[node[string, func()]]()
	s.Push(*r.root)
	for !s.IsEmpty() {
		n, ok := s.Pop()
		df = append(df, n.k.Value())
		if ok && len(n.nodes) > 0 {
			for i := len(n.nodes) - 1; i >= 0; i-- {
				s.Push(*n.nodes[i])
			}
		}
	}
	return
}

func buildHandle(key string) func() {
	return func() {
		println(key)
	}
}

func newTN(k string) (tn *testNode) {
	tn = &testNode{key: k, value: buildHandle(k)}
	return
}

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
			assert.True(t, ki.HasNext())
			assert.Equal(t, k, ki.Next())
		}
		assert.False(t, ki.HasNext())
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
		{key: "/toto/123{abc}789{def}/hello/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}, &staticKey{"/hello/"}}},
		{key: "{abc}/*", keys: []Key[string]{&regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: ":xyz/{abc}/*", keys: []Key[string]{&wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/:xyz/{abc}/*", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/:xyz/{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/123{abc}/*", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/123{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/toto/123{abc}/*/hello", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/hello"}}},
		{key: "/toto/123{abc}789/:xyz/hello/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}}},
		{key: "/toto/123{abc}789/:xyz/hello/pre*", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}}},
		{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789/:xyz/hello/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789-{\\w+}/:xyz/hello/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789-", "{\\w+}"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		{key: "123{abc}789{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		{key: "123{abc}{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		{key: "123{abc}789{(?P<date>[a-z][0-9]?)}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}}},
		{key: "123{abc}789{(?P<date>[a-z][0-9]?)}-{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}", "-", "{\\w+}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}}},
		{key: "hello-{(?P<abc>[a-z]+)}!=bonjour-{(?P<def>\\d\\w*)}-world/", keys: []Key[string]{&regexKey{value: []string{"hello-", "{(?P<abc>[a-z]+)}", "!=bonjour-", "{(?P<def>\\d\\w*)}", "-world"}, patterns: compileRePattern("{(?P<abc>[a-z]+)}", "{(?P<def>\\d\\w*)}"), params: map[string]string{"abc": "", "def": ""}}, &staticKey{"/"}}},
	}

	for _, tc := range tcs {
		ki := buildKeyIter(tc.key)
		assert.NotNil(t, ki)
		for _, k := range tc.keys {
			assert.True(t, ki.HasNext())
			assert.Equal(t, k, ki.Next())
		}
		assert.False(t, ki.HasNext())
	}
}

func TestStringKey(t *testing.T) {
	tcs := []struct {
		key      string
		path     string
		common   KeyIterator[string]
		tailKey  KeyIterator[string]
		tailPath KeyIterator[string]
	}{
		{key: "/", path: "/abc", common: buildKeyIter("/"), tailKey: nil, tailPath: buildKeyIter("abc")},
		{key: "/abc", path: "/", common: buildKeyIter("/"), tailKey: buildKeyIter("abc"), tailPath: nil},
		{key: "/abc", path: "/ab", common: buildKeyIter("/ab"), tailKey: buildKeyIter("c"), tailPath: nil},
		{key: "/123", path: "/123abc", common: buildKeyIter("/123"), tailKey: nil, tailPath: buildKeyIter("abc")},
		{key: "/123/", path: "/123/abc", common: buildKeyIter("/123/"), tailKey: nil, tailPath: buildKeyIter("abc")},
		{key: "123/", path: "/123/abc", common: nil, tailKey: buildKeyIter("123/"), tailPath: buildKeyIter("/123/abc")},
		{key: "123", path: "/123/abc", common: nil, tailKey: buildKeyIter("123"), tailPath: buildKeyIter("/123/abc")},
		{key: "123abc", path: "123/abc", common: buildKeyIter("123"), tailKey: buildKeyIter("abc"), tailPath: buildKeyIter("/abc")},
		{key: "123/abc", path: "123/", common: buildKeyIter("123/"), tailKey: buildKeyIter("abc"), tailPath: nil},
	}
	for _, tc := range tcs {
		var sk Key[string] = &staticKey{value: tc.key}
		assert.Equal(t, tc.key, sk.Value())
		c, tk, tp, _ := sk.Match(&staticKey{value: tc.path})
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
	r := New[string, func()](buildKeyIter)
	assert.Empty(t, r.root)
	assert.Empty(t, r.size)
	assert.NotNil(t, r.buildKeyIterator)
}

func TestNewNode(t *testing.T) {
	r := New[string, func()](buildKeyIter)
	ki := r.buildKeyIterator("aaa")
	n := r.newNode(ki.Next())
	assert.NotNil(t, n.k)
	assert.Nil(t, n.v)
	assert.Empty(t, n.nodes)
}

func TestPutRecWithStringKeySingleNode(t *testing.T) {
	r := New[string, func()](buildKeyIter)
	assert.True(t, r.put("/aaa", buildHandle("/aaa")))
	assert.NotNil(t, r.root)
	assert.Equal(t, 1, r.Len())
	assert.Equal(t, "/aaa", r.root.k.Value())
}

func TestPutRecWithStaticKeys(t *testing.T) {
	tcs := []struct {
		ns  []*testNode
		dfs []string
	}{
		{[]*testNode{newTN("/")}, []string{"/"}},
		{[]*testNode{newTN("/"), newTN("/abc")}, []string{"/", "abc"}},
		{[]*testNode{newTN("/"), newTN("/abc"), newTN("/abc/123")}, []string{"/", "abc", "/123"}},
		{[]*testNode{newTN("/abc"), newTN("/abc/123")}, []string{"/abc", "/123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/abc/123")}, []string{"/abc/", "def/", "123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/abc")}, []string{"/abc", "/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/")}, []string{"/", "abc/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/123")}, []string{"/", "abc/def/", "123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/abc")}, []string{"/", "abc", "/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/abc"), newTN("/abc/123")}, []string{"/", "abc", "/", "def/", "123"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def")}, []string{"/", "abc", "/def", "/hij"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def"), newTN("/abc/def/hij/123/")}, []string{"/", "abc", "/def", "/hij", "/123/"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def"), newTN("/abc/def/hij/123/"), newTN("/abc/def/hij/567")}, []string{"/", "abc", "/def", "/hij", "/", "123/", "567"}},
	}
	var r *Radix[string, func()]
	for _, tc := range tcs {
		r = New[string, func()](buildKeyIter)
		for _, i := range tc.ns {
			r.put(i.key, i.value)
		}
		result := depthFirst(r)
		assert.EqualValues(t, tc.dfs, result)
	}
}

func TestPutRecWithWildcardStar(t *testing.T) {
	tcs := []struct {
		ns  []*testNode
		dfs []string
	}{
		{[]*testNode{newTN("/")}, []string{"/"}},
		{[]*testNode{newTN("/"), newTN("/abc")}, []string{"/", "abc"}},
		{[]*testNode{newTN("/"), newTN("/abc"), newTN("/abc/123")}, []string{"/", "abc", "/123"}},
		{[]*testNode{newTN("/abc"), newTN("/abc/123")}, []string{"/abc", "/123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/abc/123")}, []string{"/abc/", "def/", "123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/abc")}, []string{"/abc", "/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/")}, []string{"/", "abc/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/123")}, []string{"/", "abc/def/", "123"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/abc")}, []string{"/", "abc", "/def/"}},
		{[]*testNode{newTN("/abc/def/"), newTN("/"), newTN("/abc"), newTN("/abc/123")}, []string{"/", "abc", "/", "def/", "123"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def")}, []string{"/", "abc", "/def", "/hij"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def"), newTN("/abc/def/hij/123/")}, []string{"/", "abc", "/def", "/hij", "/123/"}},
		{[]*testNode{newTN("/abc/def/hij"), newTN("/"), newTN("/abc"), newTN("/abc/def"), newTN("/abc/def/hij/123/"), newTN("/abc/def/hij/567")}, []string{"/", "abc", "/def", "/hij", "/", "123/", "567"}},
	}
	var r *Radix[string, func()]
	for _, tc := range tcs {
		r = New[string, func()](buildKeyIter)
		for _, i := range tc.ns {
			r.put(i.key, i.value)
		}
		result := depthFirst(r)
		assert.EqualValues(t, tc.dfs, result)
	}
}
