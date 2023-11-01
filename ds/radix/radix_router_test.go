package radix

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
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
	}
	for _, tc := range tcs {
		var sk Key[string] = &wildcardStarKey{value: tc.value}
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

func TestFormatRePattern(t *testing.T) {
	tcs := []struct {
		before string
		after  string
	}{
		{`(?P<aa>bb)`, `(bb)`},
		{`\(?P<aa>bb)`, `\(?P<aa>bb)`},
		{`\\(?P<aa>bb)`, `\\(bb)`},
		{`\\\(?P<aa>bb)`, `\\\(?P<aa>bb)`},
		{`(?P<aa>bb)cc(?P<dd>ee)ff`, `(bb)cc(ee)ff`},
		{`\(?P<aa>bb)cc(?P<dd>ee)ff`, `\(?P<aa>bb)cc(ee)ff`},
		{`\(?P<aa>bb)cc\\(?P<dd>ee)ff`, `\(?P<aa>bb)cc\\(ee)ff`},
		{`\\\(?P<aa>bb)cc\\(?P<dd>ee)ff`, `\\\(?P<aa>bb)cc\\(ee)ff`},
		{`\d`, `[0-9]`},
		{`\dabc\d`, `[0-9]abc[0-9]`},
		{`\\dabc\d`, `\\dabc[0-9]`},
		{`\\\d`, `\\[0-9]`},
		{`(?P<aa>bb)abc\d`, `(bb)abc[0-9]`},
		{`\D`, `[^0-9]`},
		{`\Dabc\D`, `[^0-9]abc[^0-9]`},
		{`\\Dabc\D`, `\\Dabc[^0-9]`},
		{`\\\D`, `\\[^0-9]`},
		{`(?P<aa>bb)abc\ddef\D`, `(bb)abc[0-9]def[^0-9]`},
		{`\x0c`, `\f`},
		{`\x0cabc\x0c`, `\fabc\f`},
		{`\\x0cabc\x0c`, `\\x0cabc\f`},
		{`\\\x0c`, `\\\f`},
		{`(?P<aa>bb)abc\ddef\D567\x0c`, `(bb)abc[0-9]def[^0-9]567\f`},
		{`\x0a`, `\n`},
		{`\x0aabc\x0a`, `\nabc\n`},
		{`\\x0aabc\x0a`, `\\x0aabc\n`},
		{`\\\x0a`, `\\\n`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0a`, `(bb)abc[0-9]def[^0-9]567\f789\f67\n`},
		{`\cJ`, `\n`},
		{`\cJabc\cJ`, `\nabc\n`},
		{`\\cJabc\cJ`, `\\cJabc\n`},
		{`\\\cJ`, `\\\n`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0ahhh\cJ`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nhhh\n`},
		{`\x0d`, `\r`},
		{`\x0dabc\x0d`, `\rabc\r`},
		{`\\x0dabc\x0d`, `\\x0dabc\r`},
		{`\\\x0d`, `\\\r`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x0d`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\r`},
		{`\cM`, `\r`},
		{`\cMabc\cM`, `\rabc\r`},
		{`\\cMabc\cM`, `\\cMabc\r`},
		{`\\\cM`, `\\\r`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0ahhh\cJcns\cM`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nhhh\ncns\r`},
		{`\s`, `[ \f\n\r\t\v]`},
		{`\sabc\s`, `[ \f\n\r\t\v]abc[ \f\n\r\t\v]`},
		{`\\sabc\s`, `\\sabc[ \f\n\r\t\v]`},
		{`\\\s`, `\\[ \f\n\r\t\v]`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x0d222\s`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\r222[ \f\n\r\t\v]`},
		{`\S`, `[^ \f\n\r\t\v]`},
		{`\Sabc\S`, `[^ \f\n\r\t\v]abc[^ \f\n\r\t\v]`},
		{`\\Sabc\S`, `\\Sabc[^ \f\n\r\t\v]`},
		{`\\\S`, `\\[^ \f\n\r\t\v]`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0ahhh\cJcns\cM999\S`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nhhh\ncns\r999[^ \f\n\r\t\v]`},
		{`\x09`, `\t`},
		{`\x09abc\x09`, `\tabc\t`},
		{`\\x09abc\x09`, `\\x09abc\t`},
		{`\\\x09`, `\\\t`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x09`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\t`},
		{`\cI`, `\t`},
		{`\cIabc\x09`, `\tabc\t`},
		{`\\cIabc\x09`, `\\cIabc\t`},
		{`\\\cI`, `\\\t`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\cI`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\t`},
		{`\x0b`, `\v`},
		{`\x0babc\x0b`, `\vabc\v`},
		{`\\x0babc\x0b`, `\\x0babc\v`},
		{`\\\x0b`, `\\\v`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x0b`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\v`},
		{`\cK`, `\v`},
		{`\cKabc\x0b`, `\vabc\v`},
		{`\\cKabc\x0b`, `\\cKabc\v`},
		{`\\\cK`, `\\\v`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\cK`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\v`},
		{`\w`, `[A-Za-z0-9_]`},
		{`\wabc\w`, `[A-Za-z0-9_]abc[A-Za-z0-9_]`},
		{`\\wabc\w`, `\\wabc[A-Za-z0-9_]`},
		{`\\\w`, `\\[A-Za-z0-9_]`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x0d222\w`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\r222[A-Za-z0-9_]`},
		{`\W`, `[^A-Za-z0-9_]`},
		{`\Wabc\W`, `[^A-Za-z0-9_]abc[^A-Za-z0-9_]`},
		{`\\Wabc\W`, `\\Wabc[^A-Za-z0-9_]`},
		{`\\\W`, `\\[^A-Za-z0-9_]`},
		{`(?P<aa>bb)abc\ddef\D567\x0c789\cL67\x0annn\x0d222\W`, `(bb)abc[0-9]def[^0-9]567\f789\f67\nnnn\r222[^A-Za-z0-9_]`},
	}
	for _, tc := range tcs {
		assert.Equal(t, tc.after, formatRePattern(tc.before))
	}
}

func TestIndexNth(t *testing.T) {
	tcs := []struct {
		key   string
		char  uint8
		occur int
		index int
	}{
		{"abcde", 's', 1, -1},
		{"abcde", 'b', 1, 1},
		{"abcbe", 'b', 2, 3},
		{"abcbb", 'b', 2, 3},
		{"abcbb", 'b', 3, 4},
		{"abcbb", 'b', 5, -1},
		{"abcbb", 'c', 2, -1},
		{"abcbb", 'c', 1, 2},
	}
	for _, tc := range tcs {
		assert.Equal(t, tc.index, indexNth(tc.key, tc.char, tc.occur))
	}
}

func TestNewKeyIter(t *testing.T) {
	tcs := []struct {
		key   string
		keys  []Key[string]
		panic string
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
		{key: "abc*", keys: []Key[string]{&staticKey{"abc"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "abc*123", panic: "Key parsing error, Invalid wildcard key: abc*123 at index: 3."},
		{key: "/*", keys: []Key[string]{&staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/abc*", keys: []Key[string]{&staticKey{"/abc"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/abc*123", panic: "Key parsing error, Invalid wildcard key: /abc*123 at index: 4."},
		{key: "*/", panic: "Key parsing error, Invalid wildcard key: */ at index: 0."},
		{key: "abc*/", panic: "Key parsing error, Invalid wildcard key: abc*/ at index: 3."},
		{key: "abc*123/", panic: "Key parsing error, Invalid wildcard key: abc*123/ at index: 3."},
		{key: "/*/", panic: "Key parsing error, Invalid wildcard key: /*/ at index: 1."},
		{key: "/123/*", keys: []Key[string]{&staticKey{"/123/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		{key: "/abc/def*", keys: []Key[string]{&staticKey{"/abc/def"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: ":abc", keys: []Key[string]{&wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}}},
		//{key: "/:abc", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}}},
		//{key: ":abc/", keys: []Key[string]{&wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		//{key: "/:abc/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		//{key: "/:abc/123/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/123/"}}},
		//{key: "/123/:abc/", keys: []Key[string]{&staticKey{"/123/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/"}}},
		//{key: "/123/:abc/789/", keys: []Key[string]{&staticKey{"/123/"}, &wildcardColonKey{value: ":abc", params: map[string]string{"abc": ""}}, &staticKey{"/789/"}}},
		//{key: "{abc}", keys: []Key[string]{&regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		//{key: "/{abc}", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		//{key: "/{abc}/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		//{key: "123{abc}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		//{key: "123{abc}789", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		//{key: "/123{abc}789", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}}},
		//{key: "123{abc}789/", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		//{key: "/123{abc}789/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}}},
		//{key: "/123{abc}789{def}/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}, &staticKey{"/"}}},
		//{key: "toto/123{abc}789{def}", keys: []Key[string]{&staticKey{"toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}}},
		//{key: "/toto/123{abc}789{def}", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}}},
		//{key: "/toto/123{abc}789{def}/hello/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789", "{def}"}, patterns: compileRePattern("{abc}", "{def}"), params: map[string]string{}}, &staticKey{"/hello/"}}},
		//{key: "{abc}/*", keys: []Key[string]{&regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: ":xyz/{abc}/*", keys: []Key[string]{&wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: "/:xyz/{abc}/*", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: "/:xyz/{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/"}, &regexKey{value: []string{"{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		//{key: "/123{abc}/*", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}}},
		//{key: "/123{abc}/*/", keys: []Key[string]{&staticKey{"/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		//{key: "/toto/123{abc}/*/hello", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardStarKey{value: "*", params: map[string]string{"*": ""}}, &staticKey{"/hello"}}},
		//{key: "/toto/123{abc}789/:xyz/hello/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}}},
		//{key: "/toto/123{abc}789/:xyz/hello/pre*", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123", "{abc}", "789"}, patterns: compileRePattern("{abc}"), params: map[string]string{}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}}},
		//{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789/:xyz/hello/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		//{key: "/toto/123-{(?P<date>[a-z][0-9]?)}-789-{\\w+}/:xyz/hello/pre*/", keys: []Key[string]{&staticKey{"/toto/"}, &regexKey{value: []string{"123-", "{(?P<date>[a-z][0-9]?)}", "-789-", "{\\w+}"}, patterns: compileRePattern("{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}, &staticKey{"/"}, &wildcardColonKey{value: ":xyz", params: map[string]string{"xyz": ""}}, &staticKey{"/hello/"}, &wildcardStarKey{value: "*", prefix: "pre", params: map[string]string{"*": ""}}, &staticKey{"/"}}},
		//{key: "123{abc}789{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		//{key: "123{abc}{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "{\\w+}"}, patterns: compileRePattern("{abc}", "{\\w+}"), params: map[string]string{}}}},
		//{key: "123{abc}789{(?P<date>[a-z][0-9]?)}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}"), params: map[string]string{"date": ""}}}},
		//{key: "123{abc}789{(?P<date>[a-z][0-9]?)}-{\\w+}", keys: []Key[string]{&regexKey{value: []string{"123", "{abc}", "789", "{(?P<date>[a-z][0-9]?)}", "-", "{\\w+}"}, patterns: compileRePattern("{abc}", "{(?P<date>[a-z][0-9]?)}", "{\\w+}"), params: map[string]string{"date": ""}}}},
		//{key: "hello-{(?P<abc>[a-z]+)}!=bonjour-{(?P<def>\\d\\w*)}-world/", keys: []Key[string]{&regexKey{value: []string{"hello-", "{(?P<abc>[a-z]+)}", "!=bonjour-", "{(?P<def>\\d\\w*)}", "-world"}, patterns: compileRePattern("{(?P<abc>[a-z]+)}", "{(?P<def>\\d\\w*)}"), params: map[string]string{"abc": "", "def": ""}}, &staticKey{"/"}}},
	}

	for _, tc := range tcs {
		if tc.panic == "" {
			ki := newKeyIter(tc.key)
			assert.NotNil(t, ki)
			for _, k := range tc.keys {
				assert.True(t, ki.HasNext())
				assert.Equal(t, k, ki.Next())
			}
			assert.False(t, ki.HasNext())
			fmt.Println(ki)
		}
	}
}

func TestToto(t *testing.T) {
	a := "abc"
	tail, ok := strings.CutPrefix(a, "abc")
	fmt.Printf("tail:%s, matched:%v", tail, ok)
}
