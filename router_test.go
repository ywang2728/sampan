package sampan

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

var (
	LruTcs = []struct {
		cap   int
		nodes []lruNode
	}{
		{cap: 0, nodes: []lruNode{}},
		{cap: 1, nodes: []lruNode{{path: "/", node: &node{part: "/"}}}},
		{cap: 3, nodes: []lruNode{{path: "/", node: &node{part: "/"}}}},
		{cap: 2, nodes: []lruNode{
			{path: "/", node: &node{part: "/"}},
			{path: "/hello", node: &node{part: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{part: "/"}},
			{path: "/hello", node: &node{part: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{part: "/"}},
			{path: "/hello", node: &node{part: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{part: "/"}},
			{path: "/hello", node: &node{part: "/hello"}},
			{path: "/world", node: &node{part: "/world"}}}},
		{cap: 2, nodes: []lruNode{
			{path: "/", node: &node{part: "/"}},
			{path: "/hello", node: &node{part: "/hello"}},
			{path: "/world", node: &node{part: "/world"}}}},
	}
)

func TestNewReDelim(t *testing.T) {
	rd := newReDelim()
	assert.NotNil(t, rd)
	assert.True(t, rd.closed())
}

func TestReDelimOpen(t *testing.T) {
	rd := newReDelim()
	assert.NotNil(t, rd)
	assert.True(t, rd.open())
	assert.False(t, rd.closed())
}

func TestReDelimClose(t *testing.T) {
	rd := newReDelim()
	assert.NotNil(t, rd)
	rd.open()
	assert.True(t, rd.close())
	assert.True(t, rd.closed())
}

func TestReDelimClosed(t *testing.T) {
	rd := newReDelim()
	assert.NotNil(t, rd)
	assert.True(t, rd.open())
	assert.False(t, rd.closed())
	assert.True(t, rd.close())
	assert.True(t, rd.closed())
}

func TestReDelimReset(t *testing.T) {
	rd := newReDelim()
	assert.NotNil(t, rd)
	assert.True(t, rd.open())
	assert.False(t, rd.closed())
	rd.reset()
	assert.True(t, rd.closed())
}

func TestNewLru(t *testing.T) {
	for _, tc := range LruTcs {
		l := newLru(tc.cap)
		assert.NotNil(t, l)
		assert.NotNil(t, l.paths)
		assert.NotNil(t, l.nodes)
	}
}

func TestLruClear(t *testing.T) {
	for _, tc := range LruTcs {
		l := newLru(tc.cap)
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		for i := 0; i < size; i++ {
			l.nodes.PushBack(&tc.nodes[i])
		}
		l.clear()
		assert.Equal(t, 0, l.len())
	}
}

func TestLruLen(t *testing.T) {
	for _, tc := range LruTcs {
		l := newLru(tc.cap)
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		for i := 0; i < size; i++ {
			ele := l.nodes.PushBack(&tc.nodes[i])
			l.paths[tc.nodes[i].path] = ele
		}
		assert.Equal(t, size, l.len())
	}
}

func TestLruPut(t *testing.T) {
	for _, tc := range LruTcs {
		var params map[string]string
		l := newLru(tc.cap)
		for _, n := range tc.nodes {
			l.put(n.path, n.node, params)
		}
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		assert.Equal(t, size, l.len())
		for i := 1; i <= size; i++ {
			assert.Equal(t, tc.nodes[len(tc.nodes)-i].node.part, l.nodes.Remove(l.nodes.Front()).(*lruNode).node.part)
		}
	}
}

func TestLruGet(t *testing.T) {
	for _, tc := range LruTcs {
		l := newLru(tc.cap)
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		for i := 0; i < size; i++ {
			ele := l.nodes.PushBack(&tc.nodes[i])
			l.paths[tc.nodes[i].path] = ele

		}
		assert.Equal(t, size, l.len())
		for i := 0; i < size; i++ {
			n, params := l.get(tc.nodes[i].path)
			assert.Equal(t, tc.nodes[i].path, n.part)
			assert.Equal(t, tc.nodes[i].path, l.nodes.Front().Value.(*lruNode).path)
			assert.Nil(t, params)
		}
	}
}

func TestLruDelete(t *testing.T) {
	for _, tc := range LruTcs {
		l := newLru(tc.cap)
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		for i := 0; i < size; i++ {
			ele := l.nodes.PushBack(&tc.nodes[i])
			l.paths[tc.nodes[i].path] = ele

		}
		for i := 0; i < size; i++ {
			l.delete(tc.nodes[i].path)
		}
		assert.Equal(t, 0, l.len())
	}
}

func TestLcp(t *testing.T) {
	tcs := []struct {
		s1  string
		s2  string
		idx int
		cp  string
	}{
		{s1: "/", s2: "1", idx: -1, cp: ""},
		{s1: "/", s2: "/", idx: 0, cp: "/"},
		{s1: "/1", s2: "1", idx: -1, cp: ""},
		{s1: "/1", s2: "/1", idx: 0, cp: "/1"},
		{s1: "/1/", s2: "/2/", idx: 0, cp: "/"},
		{s1: "/1/", s2: "/1/", idx: 2, cp: "/1/"},
		{s1: "/1/2/", s2: "/1/2", idx: 2, cp: "/1/2"},
		{s1: "/", s2: "/", idx: 0, cp: "/"},
		{s1: "/", s2: "/123", idx: 0, cp: "/"},
		{s1: "/123", s2: "/", idx: 0, cp: "/"},
		{s1: "/12/456", s2: "/12/567", idx: 3, cp: "/12/"},
		{s1: "/12/456", s2: "/12/456/", idx: 3, cp: "/12/456"},
		{s1: "/:123/456", s2: "/*123/567", idx: 0, cp: "/"},
		{s1: "/123/4/56", s2: "/123/4/57", idx: 6, cp: "/123/4/"},
		{s1: "/123/4/5/6", s2: "/123/4/5/7", idx: 8, cp: "/123/4/5/"},
		{s1: "/123/4/56/", s2: "/123/4/56", idx: 6, cp: "/123/4/56"},
		{s1: "/123/4/56/", s2: "/123/4/56/", idx: 9, cp: "/123/4/56/"},
		{s1: "/123/4/56", s2: "/123/4/56", idx: 6, cp: "/123/4/56"},
	}
	for _, tc := range tcs {
		cp, idx := commonPrefix(tc.s1, tc.s2)
		assert.Equal(t, tc.idx, idx)
		assert.Equal(t, tc.cp, cp)
	}
}

func TestParsePref(t *testing.T) {
	tcs := []struct {
		path string
		idx  int
		eof  bool
	}{
		{path: "/", idx: 0, eof: true},
		{path: "/123", idx: 3, eof: true},
		{path: "/123/", idx: 4, eof: true},
		{path: "123/", idx: 3, eof: true},
		{path: "123", idx: 2, eof: true},
		{path: "/{abc}", idx: 0, eof: false},
		{path: "/{abc}/", idx: 0, eof: false},
		{path: "{abc}/", idx: 5, eof: true},
		{path: "{abc}", idx: 4, eof: true},
		{path: "/123{abc}", idx: 0, eof: false},
		{path: "/123{abc}/", idx: 0, eof: false},
		{path: "123{abc}/", idx: 8, eof: true},
		{path: "123{abc}", idx: 7, eof: true},
		{path: "/{abc}/123", idx: 0, eof: false},
		{path: "/{abc}/123/", idx: 0, eof: false},
		{path: "{abc}/123", idx: 5, eof: false},
		{path: "{abc}/123/", idx: 5, eof: false},
		{path: "/123{abc}/123", idx: 0, eof: false},
		{path: "/123{abc}/123/", idx: 0, eof: false},
		{path: "123{abc}/123", idx: 8, eof: false},
		{path: "123{abc}/123/", idx: 8, eof: false},
		{path: "/123/{abc}", idx: 4, eof: false},
		{path: "/123/{abc}/", idx: 4, eof: false},
		{path: "123/{abc}", idx: 3, eof: false},
		{path: "123/{abc}/", idx: 3, eof: false},
		{path: "/123/123{abc}", idx: 4, eof: false},
		{path: "/123/123{abc}/", idx: 4, eof: false},
		{path: "123/123{abc}", idx: 3, eof: false},
		{path: "123/123{abc}/", idx: 3, eof: false},
	}
	for _, tc := range tcs {
		idx, eof := parsePref(tc.path)
		assert.Equal(t, tc.idx, idx)
		assert.Equal(t, tc.eof, eof)
	}
}

func TestParseKey(t *testing.T) {
	tcs := []struct {
		path string
		key  string
	}{
		{path: "/", key: "/"},
		{path: "/123", key: "/"},
		{path: "/123/", key: "/"},
		{path: "123/", key: "123/"},
		{path: "123", key: "123"},
		{path: "/{abc}", key: "/"},
		{path: "/{abc}/", key: "/"},
		{path: "{abc}/", key: "{abc}/"},
		{path: "{abc}", key: "{abc}"},
		{path: "/123{abc}", key: "/"},
		{path: "/123{abc}/", key: "/"},
		{path: "123{abc}/", key: "123{abc}/"},
		{path: "123{abc}", key: "123{abc}"},
		{path: "/{abc}/123", key: "/"},
		{path: "/{abc}/123/", key: "/"},
		{path: "{abc}/123", key: "{abc}/"},
		{path: "{abc}/123/", key: "{abc}/"},
		{path: "/123{abc}/123", key: "/"},
		{path: "/123{abc}/123/", key: "/"},
		{path: "123{abc}/123", key: "123{abc}/"},
		{path: "123{abc}/123/", key: "123{abc}/"},
		{path: "/123/{abc}", key: "/"},
		{path: "/123/{abc}/", key: "/"},
		{path: "123/{abc}", key: "123/"},
		{path: "123/{abc}/", key: "123/"},
		{path: "/123/123{abc}", key: "/"},
		{path: "/123/123{abc}/", key: "/"},
		{path: "123/123{abc}", key: "123/"},
		{path: "123/123{abc}/", key: "123/"},
	}
	for _, tc := range tcs {
		key := parseKey(tc.path)
		assert.Equal(t, tc.key, key)
	}
}

func TestParseRe(t *testing.T) {
	tcs := []struct {
		part string
		exps []string
		cnt  int
	}{
		{part: "{abc}/", exps: []string{"abc"}, cnt: 1},
		{part: "{abc:abc}/", exps: []string{"abc:abc"}, cnt: 1},
		{part: "hello-{abc}-{def}haha/", exps: []string{"abc", "def"}, cnt: 2},
		{part: "hello-{abc:ccc}-{def}haha/", exps: []string{"abc:ccc", "def"}, cnt: 2},
		{part: "{ab::c}/", exps: []string{"ab::c"}, cnt: 1},
		{part: "{abc:[a-z]+}/", exps: []string{"abc:[a-z]+"}, cnt: 1},
		{part: "{abc:[a-z]{3-5}}/", exps: []string{"abc:[a-z]{3-5}"}, cnt: 1},
		{part: "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/", exps: []string{"abc:[a-z]{3-5}", "haha:\\w+"}, cnt: 2},
		{part: "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/", exps: []string{"abc:[a-z]{3-5}", "haha:\\w+", "a1:[0-9][0-9]?"}, cnt: 3},
	}
	for _, tc := range tcs {
		assert.Equal(t, len(parseRe(tc.part)), tc.cnt)
		assert.Equal(t, parseRe(tc.part), tc.exps)
	}
}

func TestParseRePattern(t *testing.T) {
	tcs := []struct {
		part string
		bef  string
		aft  string
		isRe bool
	}{
		{part: "/", bef: "/", aft: "", isRe: false},
		{part: "hello/", bef: "hello/", aft: "", isRe: false},
		{part: "hello{name}/", bef: "hello", aft: "{name}/", isRe: false},
		{part: "{abc:abc}/", bef: "{abc:abc}", aft: "/", isRe: true},
		{part: "{ab::c}/", bef: "{ab::c}", aft: "/", isRe: true},
		{part: "hello-{abc}-{def}haha/", bef: "hello-", aft: "{abc}-{def}haha/", isRe: false},
		{part: "{abc:[a-z]+}/", bef: "{abc:[a-z]+}", aft: "/", isRe: true},
		{part: "{abc}", bef: "{abc}", aft: "", isRe: true},
		{part: "{abc}/", bef: "{abc}", aft: "/", isRe: true},
		{part: "{abc}world", bef: "{abc}", aft: "world", isRe: true},
		{part: "{abc}world/", bef: "{abc}", aft: "world/", isRe: true},
		{part: "{abc}-world/", bef: "{abc}", aft: "-world/", isRe: true},
		{part: "{abc:[a-z]{3-5}}/", bef: "{abc:[a-z]{3-5}}", aft: "/", isRe: true},
		{part: "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/", bef: "123-", aft: "{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/", isRe: false},
		{part: "{abc:[a-z]{3-5}}-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/", bef: "{abc:[a-z]{3-5}}", aft: "-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/", isRe: true},
	}
	for _, tc := range tcs {
		bef, aft, isRe := parseRePattern(tc.part)
		assert.Equal(t, tc.bef, bef)
		assert.Equal(t, tc.aft, aft)
		assert.Equal(t, tc.isRe, isRe)
	}
}

func TestNewReMap(t *testing.T) {
	tcs := []struct {
		part string
		divs []string
		exps map[string]string
	}{
		{part: "/", divs: []string{"/"}, exps: map[string]string{}},
		{part: "/abc", divs: []string{"/abc"}, exps: map[string]string{}},
		{part: "abc", divs: []string{"abc"}, exps: map[string]string{}},
		{part: "abc/", divs: []string{"abc/"}, exps: map[string]string{}},
		{part: "{abc}", divs: []string{"{abc}"}, exps: map[string]string{"abc": `\S*`}},
		{part: "/{abc}", divs: []string{"/", "{abc}"}, exps: map[string]string{"abc": `\S*`}},
		{part: "{abc}/", divs: []string{"{abc}", "/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "/{abc}/", divs: []string{"/", "{abc}", "/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "{abc}world", divs: []string{"{abc}", "world"}, exps: map[string]string{"abc": `\S*`}},
		{part: "/{abc}world", divs: []string{"/", "{abc}", "world"}, exps: map[string]string{"abc": `\S*`}},
		{part: "{abc}world/", divs: []string{"{abc}", "world/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "/{abc}world/", divs: []string{"/", "{abc}", "world/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "{abc}/world/", divs: []string{"{abc}", "/world/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "/{abc}/world/", divs: []string{"/", "{abc}", "/world/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "hello{name}/", divs: []string{"hello", "{name}", "/"}, exps: map[string]string{"name": `\S*`}},
		{part: "{abc:abc}/", divs: []string{"{abc:abc}", "/"}, exps: map[string]string{"abc": `abc`}},
		{part: "{ab::c}/", divs: []string{"{ab::c}", "/"}, exps: map[string]string{"ab": `:c`}},
		{
			part: "hello-{abc}-{def}haha/",
			divs: []string{"hello-", "{abc}", "-", "{def}", "haha/"},
			exps: map[string]string{"abc": `\S*`, "def": `\S*`},
		},
		{part: "{abc:[a-z]+}/", divs: []string{"{abc:[a-z]+}", "/"}, exps: map[string]string{"abc": `[a-z]+`}},
		{part: "{abc:[a-z]{3-5}}/", divs: []string{"{abc:[a-z]{3-5}}", "/"}, exps: map[string]string{"abc": `[a-z]{3-5}`}},
		{
			part: "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/",
			divs: []string{"123-", "{abc:[a-z]{3-5}}", "-567-", "{haha:\\w+}", "--world/"},
			exps: map[string]string{"abc": `[a-z]{3-5}`, "haha": `\w+`},
		},
		{
			part: "{abc:[a-z]{3-5}}-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/",
			divs: []string{"{abc:[a-z]{3-5}}", "-567-", "{haha:\\w+}", "--world++", "{a1:[0-9][0-9]?}", "ll/"},
			exps: map[string]string{"abc": `[a-z]{3-5}`, "haha": `\w+`, "a1": `[0-9][0-9]?`},
		},
	}
	for _, tc := range tcs {
		rm := newReMap(tc.part)
		assert.NotNil(t, rm)
		assert.Equal(t, len(tc.divs), rm.len())
		for _, v := range rm.divs {
			assert.Contains(t, tc.divs, v)
		}
		for k, v := range rm.exps {
			v2, ok := tc.exps[k]
			assert.True(t, ok)
			assert.Equal(t, v2, v.String())
		}
	}

}

func TestNewNode(t *testing.T) {
	tcs := []struct {
		part string
		divs []string
		exps map[string]string
	}{
		{part: "/", divs: nil, exps: nil},
		{part: "abc/", divs: nil, exps: nil},
		{part: "{abc}/", divs: []string{"{abc}", "/"}, exps: map[string]string{"abc": `\S*`}},
		{part: "{abc:\\w+}/", divs: []string{"{abc:\\w+}", "/"}, exps: map[string]string{"abc": `\w+`}},
		{part: "{abc:[a-z][0-9]?}/", divs: []string{"{abc:[a-z][0-9]?}", "/"}, exps: map[string]string{"abc": `[a-z][0-9]?`}},
		{part: "hello-{abc:[a-z]+}!/", divs: []string{"hello-", "{abc:[a-z]+}", "!/"}, exps: map[string]string{"abc": `[a-z]+`}},
		{
			part: "hello-{abc:[a-z]+}!=bonjour-{def:[\\d][\\w]*}/",
			divs: []string{"hello-", "{abc:[a-z]+}", "!=bonjour-", "{def:[\\d][\\w]*}", "/"},
			exps: map[string]string{"abc": `[a-z]+`, "def": "[\\d][\\w]*"}},
	}
	for _, tc := range tcs {
		n := newNode(tc.part)
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
		if tc.exps == nil {
			assert.Nil(t, n.rePatterns)
		} else {
			for _, v := range n.rePatterns.divs {
				assert.Contains(t, tc.divs, v)
			}
			assert.Equal(t, len(n.rePatterns.exps), len(tc.exps))
			for k, v := range n.rePatterns.exps {
				v2, ok := tc.exps[k]
				assert.True(t, ok)
				assert.Equal(t, v2, v.String())
			}
		}
		assert.NotNil(t, n.children)
		assert.Nil(t, n.handler)
	}
}

func TestNewRadix(t *testing.T) {
	r := newRadix()
	assert.NotNil(t, r)
}

func TestRadixClear(t *testing.T) {
	r := newRadix()
	r.root = newNode("/")
	r.size++
	assert.Equal(t, 1, r.len())
	r.clear()
	assert.Equal(t, 0, r.len())
}

func TestRadixLen(t *testing.T) {
	r := newRadix()
	r.root = newNode("/")
	r.size++
	assert.Equal(t, 1, r.len())
	r.root.children["hello"] = newNode("hello")
	r.size++
	assert.Equal(t, 2, r.len())
}

func TestRadixString(t *testing.T) {
	r := newRadix()
	r.root = newNode("/")
	r.root.children["hello"] = newNode("hello")
	r.size = 2
	fmt.Println(r)
}

func TestRadixPutRecNewSinglePlainTextPath(t *testing.T) {
	tcs := []struct {
		path string
		part string
	}{
		{path: "/", part: "/"},
		{path: "/123", part: "/123"},
		{path: "/123/", part: "/123/"},
		{path: "123/", part: "123/"},
	}
	for _, tc := range tcs {
		r := newRadix()
		n := r.putRec(nil, tc.path, func(context *Context) {})
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
	}
}

func TestRadixPutRecNewSingleRegexPath(t *testing.T) {
	tcs := []struct {
		path      string
		part      string
		reDivs    []string
		reExpsKey string
		reExpsStr string
	}{
		{path: "{abc}/", part: "{abc}/", reDivs: []string{"{abc}", "/"}, reExpsKey: "abc", reExpsStr: `\S*`},
		{path: "123{abc}/", part: "123{abc}/", reDivs: []string{"123", "{abc}", "/"}, reExpsKey: "abc", reExpsStr: `\S*`},
	}
	for _, tc := range tcs {
		r := newRadix()
		n := r.putRec(nil, tc.path, func(context *Context) {})
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
		assert.Equal(t, len(tc.reDivs), n.rePatterns.len())
		assert.Equal(t, 1, len(n.rePatterns.exps))
		assert.Equal(t, tc.reExpsStr, n.rePatterns.exps[tc.reExpsKey].String())
	}
}

func TestRadixPutRecNewMultipleRegexPath(t *testing.T) {
	tcs := []struct {
		path  string
		part  string
		nodes []map[string]string
	}{
		{path: "/123/{abc}", part: "/123/", nodes: []map[string]string{{"part": "{abc}/", "expKey": "abc", "expStr": `\S*`}}},
		{path: "/123/{abc}/def/", part: "/123/", nodes: []map[string]string{{"part": "{abc}/", "expKey": "abc", "expStr": `\S*`}}},
		{path: "/123/{abc}/456/", part: "/123/", nodes: []map[string]string{{"part": "{abc}/", "expKey": "abc", "expStr": `\S*`}}},
	}
	for _, tc := range tcs {
		r := newRadix()
		handler := func(ctx *Context) {}
		n := r.putRec(nil, tc.path, handler)
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
		r.root = n
		fmt.Println(r)
	}
}

func TestRadixPutRecPlainTextPaths(t *testing.T) {
	tcs := []struct {
		path    string
		part    string
		handler func(ctx *Context)
		nodes   []map[string]string
	}{
		{path: "/123/", part: "/123/", handler: func(ctx *Context) {}},
		{path: "/", part: "/", handler: func(ctx *Context) {}},
		{path: "/abc/", part: "/abc/", handler: func(ctx *Context) {}},
		{path: "/1234/", part: "1234/", handler: func(ctx *Context) {}},
		{path: "/123/abc", part: "abc", handler: func(ctx *Context) {}},
		{path: "/123/def/hij", part: "def/", handler: func(ctx *Context) {}},
		{path: "/123/def/ghi/", part: "ghi/", handler: func(ctx *Context) {}},
		{path: "/123/def/ghi", part: "ghi/", handler: func(ctx *Context) {}},
		{path: "/123/def/haha", part: "ghi/56789", handler: func(ctx *Context) {}},
		{path: "/123/def/haha/99999/8888", part: "ghi/56789", handler: func(ctx *Context) {}},
		{path: "/123/def/haha/56789/6666/", part: "ghi/56789", handler: func(ctx *Context) {}},
	}
	r := newRadix()
	for _, tc := range tcs {
		var n *node
		if n = r.putRec(r.root, tc.path, tc.handler); n != nil {
			r.root = n
			r.size++
		}
		assert.NotNil(t, n)
	}
	assert.NotNil(t, r)
	assert.Equal(t, len(tcs), r.len())
	fmt.Println(r)
}

func TestRadixPutRecMixedPaths(t *testing.T) {
	tcs := []struct {
		path    string
		part    string
		handler func(ctx *Context)
		nodes   []map[string]string
	}{
		{path: "/{abc}/", part: "/{abc}/", handler: func(ctx *Context) {}},
		{path: "/{abc}/123", part: "{def}", handler: func(ctx *Context) {}},
		{path: "/{abc}/456/789", part: "{def}", handler: func(ctx *Context) {}},
		{path: "/{abc}/{toto}", part: "/abc/", handler: func(ctx *Context) {}},
		{path: "/{abc}/{toto}/", part: "/abc/", handler: func(ctx *Context) {}},
		{path: "/{abc}/{toto}/haha", part: "/abc/", handler: func(ctx *Context) {}},
		{path: "/{abc}/{toto}/{nini}/", part: "/abc/", handler: func(ctx *Context) {}},
	}
	r := newRadix()
	for _, tc := range tcs {
		var n *node
		if n = r.putRec(r.root, tc.path, tc.handler); n != nil {
			r.root = n
			r.size++
		}
		assert.NotNil(t, n)
	}
	assert.NotNil(t, r)
	assert.Equal(t, len(tcs), r.len())
	fmt.Println(r)
}

func TestGetRecPlainTextPath(t *testing.T) {
	tcs := []struct {
		path    string
		part    string
		handler func(ctx *Context)
		nodes   []map[string]string
	}{
		{path: "/123/toto", part: "/123/", handler: func(ctx *Context) { fmt.Println("func:", "/123/toto") }},
		{path: "/abc/def", part: "/abc/", handler: func(ctx *Context) { fmt.Println("func:", "/abc/def") }},
		{path: "/123/haha/", part: "1234/", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha/") }},
		{path: "/123/haha", part: "1234/", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha") }},
		{path: "/123/haha/nini", part: "1234/", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha/nini") }},
	}
	r := newRadix()
	for _, tc := range tcs {
		var n *node
		if n = r.putRec(r.root, tc.path, tc.handler); n != nil {
			r.root = n
			r.size++
		}
		assert.NotNil(t, n)
	}
	assert.NotNil(t, r)
	assert.Equal(t, len(tcs), r.len())
	fmt.Println(r)
	for _, tc := range tcs {
		var params map[string]string
		n := r.getRec(r.root, tc.path, params)
		if n != nil {
			n.handler(nil)
		} else {
			fmt.Println("n is nil")
		}
	}
}

func TestNewRouter(t *testing.T) {
	r := newRouter()
	assert.NotNil(t, r)
	assert.NotNil(t, r.trees)
	fmt.Println(strings.CutPrefix("abc", "ab"))
}
