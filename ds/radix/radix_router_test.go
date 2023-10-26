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

// staticKey
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

func TestStaticKeyMatch(t *testing.T) {
	tcs := []struct {
		value   string
		path    string
		tail    string
		params  map[string]string
		matched bool
	}{
		{value: "/", path: "/abc", tail: "abc", params: map[string]string{}, matched: true},
		{value: "/abc", path: "/", tail: "/", params: map[string]string{}, matched: false},
		{value: "/abc", path: "/ab", tail: "/ab", params: map[string]string{}, matched: false},
		{value: "/123", path: "/123abc", tail: "abc", params: map[string]string{}, matched: true},
		{value: "/123/", path: "/123/abc", tail: "abc", params: map[string]string{}, matched: true},
		{value: "123/", path: "/123/abc", tail: "/123/abc", params: map[string]string{}, matched: false},
		{value: "123", path: "/123/abc", tail: "/123/abc", params: map[string]string{}, matched: false},
		{value: "123abc", path: "123/abc", tail: "123/abc", params: map[string]string{}, matched: false},
		{value: "123/abc", path: "123/", tail: "123/", params: map[string]string{}, matched: false},
	}
	for _, tc := range tcs {
		var sk Key[string] = &staticKey{value: tc.value}
		tt, p, m := sk.Match(tc.path)
		assert.Equal(t, tc.tail, tt)
		assert.Equal(t, len(tc.params), len(p))
		assert.Equal(t, tc.matched, m)

	}
}

// wildcardStarKey
func TestWildcardStarKeyString(t *testing.T) {
	tcs := []struct {
		value  string
		prefix string
		suffix string
		output string
	}{
		{value: "*", output: "*"},
		{value: "*", prefix: "abc", output: "abc*"},
		{value: "*", suffix: "def", output: "*def"},
		{value: "*", prefix: "abc", suffix: "def", output: "abc*def"},
		{value: "*", prefix: "abc", suffix: "def", output: "abc*def"},
		{value: "*", prefix: "abc", suffix: "123/def", output: "abc*123/def"},
		{value: "*", prefix: "abc/123", suffix: "def", output: "abc/123*def"},
		{value: "*", prefix: "abc/123", suffix: "789/def", output: "abc/123*789/def"},
	}
	for _, tc := range tcs {
		var sk Key[string] = &wildcardStarKey{value: tc.value, prefix: tc.prefix, suffix: tc.suffix}
		assert.Equal(t, tc.output, sk.String())
		assert.Equal(t, tc.output, fmt.Sprint(sk))
	}
}

func TestWildcardStarKeyMatchIterator(t *testing.T) {
	tcs := []struct {
		value    string
		prefix   string
		suffix   string
		path     KeyIterator[string]
		common   KeyIterator[string]
		tailKey  KeyIterator[string]
		tailPath KeyIterator[string]
	}{
		{value: "*", path: buildStaticKeyIter("abc"), common: buildStaticKeyIter("/"), tailKey: nil, tailPath: buildStaticKeyIter("abc")},
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

// KeySeparator
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

func TestKeySeparatorOpenedAndClosed(t *testing.T) {
	tcs := []struct {
		bgn    string
		end    string
		char   string
		opened bool
		closed bool
	}{
		{bgn: "(", end: ")", char: "(", opened: true, closed: false},
		{bgn: ":", end: "/", char: "(", opened: false, closed: true},
		{bgn: "{", end: "}", char: "(", opened: false, closed: true},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		assert.False(t, ks.opened())
		assert.True(t, ks.closed())
		ks.openWith(tc.char)
		assert.Equal(t, tc.opened, ks.opened())
		assert.Equal(t, tc.closed, ks.closed())
	}
}

func TestKeySeparatorOpenWithAndCloseWith(t *testing.T) {
	tcs := []struct {
		bgn         string
		end         string
		bgnChars    []string
		endChars    []string
		closedTimes int
		opened      bool
		closed      bool
	}{
		{bgn: "(", end: ")", bgnChars: []string{"(", "("}, endChars: []string{")", ")"}, closedTimes: 0, opened: true, closed: true},
		{bgn: ":", end: "/", bgnChars: []string{":"}, endChars: []string{"/"}, closedTimes: 0, opened: true, closed: true},
		{bgn: "{", end: "}", bgnChars: []string{"{", "{"}, endChars: []string{"}", ")"}, closedTimes: 1, opened: true, closed: false},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		var times int
		var status bool
		for _, c := range tc.bgnChars {
			times, status = ks.openWith(c)
		}
		assert.Equal(t, len(tc.bgnChars), times)
		assert.Equal(t, tc.opened, status)
		for _, c := range tc.endChars {
			times, status = ks.closeWith(c)
		}
		assert.Equal(t, tc.closedTimes, times)
		assert.Equal(t, tc.closed, status)
	}
}

func TestKeySeparatorOpenAndClose(t *testing.T) {
	tcs := []struct {
		bgn           string
		end           string
		bgnChars      []string
		endChars      []string
		closedTimes   int
		opened        bool
		closed        bool
		isForceClosed bool
	}{
		{bgn: "{", end: "}", bgnChars: []string{"{", "{"}, endChars: []string{"}", ")"}, closedTimes: 0, opened: true, closed: false, isForceClosed: true},
	}
	for _, tc := range tcs {
		ks := newKeySeparator(tc.bgn, tc.end)
		var times int
		var status bool
		for range tc.bgnChars {
			times, status = ks.open()
		}
		assert.Equal(t, len(tc.bgnChars), times)
		assert.Equal(t, tc.opened, status)
		for _, c := range tc.endChars {
			times, status = ks.closeWith(c)
		}
		assert.Equal(t, tc.closed, status)
		times, status = ks.close()
		assert.Equal(t, tc.closedTimes, times)
		assert.Equal(t, tc.isForceClosed, status)
		assert.Equal(t, tc.isForceClosed, ks.closed())
	}
}

func TestToto(t *testing.T) {
	type Hello struct {
		s []string
		c int
	}
	a := Hello{[]string{"abc"}, 10}
	b := a
	b.c = 20
	b.s = append(b.s, "123")
	fmt.Printf("%+v  slice p: %p %p\n", a, &a.c, &a)
	fmt.Printf("%+v slice p: %p %p\n", b, &b.c, &b)
}
