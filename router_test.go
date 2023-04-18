package sampan

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"regexp"
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
		idx, eof := parsePrefix(tc.path)
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

func TestParseRePatterns(t *testing.T) {
	tcs := []struct {
		part     string
		patterns []rePattern
	}{
		{"/", []rePattern{{"/", nil}}},
		{"/abc", []rePattern{{"/abc", nil}}},
		{"abc/", []rePattern{{"abc/", nil}}},
		{"/abc/", []rePattern{{"/abc/", nil}}},
		{"/abc/def", []rePattern{{"/abc/def", nil}}},
		{"abc/def/", []rePattern{{"abc/def/", nil}}},
		{"/abc/def/", []rePattern{{"/abc/def/", nil}}},
		{"{abc}", []rePattern{{"{abc}", regexp.MustCompile("abc")}}},
		{"/{abc}", []rePattern{{"/", nil}, {"{abc}", regexp.MustCompile("abc")}}},
		{"{abc}/", []rePattern{{"{abc}", regexp.MustCompile("abc")}, {"/", nil}}},
		{"/{abc}/", []rePattern{{"/", nil}, {"{abc}", regexp.MustCompile("abc")}, {"/", nil}}},
		{"123{abc}", []rePattern{{"123", nil}, {"{abc}", regexp.MustCompile("abc")}}},
		{"/123{abc}", []rePattern{{"/123", nil}, {"{abc}", regexp.MustCompile("abc")}}},
		{"123{abc}/", []rePattern{{"123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"/", nil}}},
		{"/123{abc}/", []rePattern{{"/123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"/", nil}}},
		{"123{abc}789", []rePattern{{"123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"789", nil}}},
		{"/123{abc}789", []rePattern{{"/123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"789", nil}}},
		{"123{abc}789/", []rePattern{{"123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"789/", nil}}},
		{"/123{abc}789/", []rePattern{{"/123", nil}, {"{abc}", regexp.MustCompile("abc")}, {"789/", nil}}},
		{"{[1-9]+}{[a-z]+}", []rePattern{{"{[1-9]+}", regexp.MustCompile("[1-9]+")}, {"{[a-z]+}", regexp.MustCompile("[a-z]+")}}},
		{"/{[1-9]+}{[a-z]+}", []rePattern{{"/", nil}, {"{[1-9]+}", regexp.MustCompile("[1-9]+")}, {"{[a-z]+}", regexp.MustCompile("[a-z]+")}}},
		{"{[1-9]+}{[a-z]+}/", []rePattern{{"{[1-9]+}", regexp.MustCompile("[1-9]+")}, {"{[a-z]+}", regexp.MustCompile("[a-z]+")}, {"/", nil}}},
		{"/{[1-9]+}{[a-z]+}/", []rePattern{{"/", nil}, {"{[1-9]+}", regexp.MustCompile("[1-9]+")}, {"{[a-z]+}", regexp.MustCompile("[a-z]+")}, {"/", nil}}},
	}
	for _, tc := range tcs {
		rePatterns := parseRePatterns(tc.part)
		assert.Equal(t, len(tc.patterns), len(rePatterns))
		for i := 0; i < len(rePatterns); i++ {
			assert.Equal(t, tc.patterns[i].raw, rePatterns[i].raw)
			if rePatterns[i].compiled != nil {
				assert.Equal(t, tc.patterns[i].compiled.String(), rePatterns[i].compiled.String())
				//fmt.Printf("{raw: %s, compiled: %s}\n", rePatterns[i].raw, rePatterns[i].compiled.String())
			} else {
				//fmt.Printf("{raw: %s, compiled: %v}\n", rePatterns[i].raw, nil)
			}
		}
		//println()
	}
}

func TestNewNode(t *testing.T) {
	tcs := []struct {
		part       string
		rePatterns []rePattern
	}{
		{"/", nil},
		{"abc/", nil},
		{"{abc}/", []rePattern{{"{abc}", regexp.MustCompile(`abc`)}, {"/", nil}}},
		{"{\\w+}/", []rePattern{{"{\\w+}", regexp.MustCompile(`\w+`)}, {"/", nil}}},
		{"{(?P<date>[a-z][0-9]?)}/", []rePattern{{"{(?P<date>[a-z][0-9]?)}", regexp.MustCompile(`(?P<date>[a-z][0-9]?)`)}, {"/", nil}}},
		{"hello-{(?P<abc>[a-z]+)}!/", []rePattern{{"hello-", nil}, {"{(?P<abc>[a-z]+)}", regexp.MustCompile(`(?P<abc>[a-z]+)`)}, {"!/", nil}}},
		{
			"hello-{(?P<abc>[a-z]+)}!=bonjour-{(?P<def>\\d\\w*)}-world/",
			[]rePattern{
				{"hello-", nil},
				{"{(?P<abc>[a-z]+)}", regexp.MustCompile(`(?P<abc>[a-z]+)`)},
				{"!=bonjour-", nil},
				{"{(?P<def>\\d\\w*)}", regexp.MustCompile(`(?P<def>\d\w*)`)},
				{"-world/", nil}}},
	}
	for _, tc := range tcs {
		n := newNode(tc.part)
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
		if tc.rePatterns == nil {
			assert.Nil(t, n.rePatterns)
		} else {
			for i := 0; i < len(n.rePatterns); i++ {
				assert.Equal(t, tc.rePatterns[i].raw, n.rePatterns[i].raw)
				if n.rePatterns[i].compiled != nil {
					assert.Equal(t, tc.rePatterns[i].compiled.String(), n.rePatterns[i].compiled.String())
				}
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
		path       string
		part       string
		rePatterns []rePattern
	}{
		{"{abc}/", "{abc}/", []rePattern{{"{abc}", regexp.MustCompile(`abc`)}, {"/", nil}}},
		{"123{abc}/", "123{abc}/", []rePattern{{"123", nil}, {"{abc}", regexp.MustCompile(`abc`)}, {"/", nil}}},
	}
	for _, tc := range tcs {
		r := newRadix()
		n := r.putRec(nil, tc.path, func(context *Context) {})
		assert.NotNil(t, n)
		assert.Equal(t, tc.part, n.part)
		assert.Equal(t, len(tc.rePatterns), len(n.rePatterns))
		for i := 0; i < len(n.rePatterns); i++ {
			assert.Equal(t, tc.rePatterns[i].raw, n.rePatterns[i].raw)
			if n.rePatterns[i].compiled != nil {
				assert.Equal(t, tc.rePatterns[i].compiled.String(), n.rePatterns[i].compiled.String())
			}
		}

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
		url     string
		params  map[string]string
		handler func(ctx *Context)
		nodes   []map[string]string
	}{
		{path: "/123/", url: "/123/", handler: func(ctx *Context) { fmt.Println("func:", "/123/") }},
		//{path: "/abc/def", url: "/abc/def", handler: func(ctx *Context) { fmt.Println("func:", "/abc/def") }},
		//{path: "/123/haha/", url: "/123/haha/", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha/") }},
		{path: "/123/haha", url: "/123/haha", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha") }},
		{path: "/123/haha/nini", url: "/123/haha/nini", handler: func(ctx *Context) { fmt.Println("func:", "/123/haha/nini") }},
		//{path: "/123/{hello[0-9]{1,3}}", url: "/123/hello123", handler: func(ctx *Context) { fmt.Println("func:", "/123/{hello[0-9]{1,3}}") }},
		//{path: "/123/{hello[0-9]{1,3}}abc", url: "/123/hello123abc", handler: func(ctx *Context) { fmt.Println("func:", "/123/{hello[0-9]{1,3}}abc") }},
		//{path: "/123/{hello[A-Z]{1,3}}", url: "/123/helloABC", handler: func(ctx *Context) { fmt.Println("func:", "/123/{hello[A-Z]{1,3}}") }},
		//{path: "/123/{(?P<v1>hello[0-9]{1,3})}", url: "/123/hello123", params: map[string]string{"v1": "hello123"}, handler: func(ctx *Context) { fmt.Println("func:", "/123/{(?P<v1>hello[0-9]{1,3})}") }},
		//{path: "/123/{hello[0-9]{1,3}}/pig", url: "/123/hello123/pig", handler: func(ctx *Context) { fmt.Println("func:", "/123/{hello[0-9]{1,3}}/pig") }},
		//{path: "/123/{(?P<v1>hello[0-9]{1,3})}/pig", url: "/123/hello123/pig", params: map[string]string{"v1": "hello123"}, handler: func(ctx *Context) { fmt.Println("func:", "/123/{(?P<v1>hello[0-9]{1,3})}/pig") }},
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
	var cnt []*node
	for _, tc := range tcs {
		params := make(map[string]string)
		n := r.getRec(r.root, tc.url, params)
		if n != nil {
			cnt = append(cnt, n)
			n.handler(nil)
			if tc.params != nil {
				fmt.Printf("%+v\n", params)
				for k, v := range tc.params {
					assert.Equal(t, v, params[k])
				}
			}
		} else {
			fmt.Println("n is nil")
		}
	}
	assert.Equal(t, len(tcs), len(cnt))
}

func TestNewRouter(t *testing.T) {
	r := newRouter()
	assert.NotNil(t, r)
	assert.NotNil(t, r.trees)
	fmt.Println(strings.CutPrefix("", "abc"))
}
