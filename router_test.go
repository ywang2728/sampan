package sampan

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

var (
	LruTcs = []struct {
		cap   int
		nodes []lruNode
	}{
		{cap: 0, nodes: []lruNode{}},
		{cap: 1, nodes: []lruNode{{path: "/", node: &node{path: "/"}}}},
		{cap: 3, nodes: []lruNode{{path: "/", node: &node{path: "/"}}}},
		{cap: 2, nodes: []lruNode{
			{path: "/", node: &node{path: "/"}},
			{path: "/hello", node: &node{path: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{path: "/"}},
			{path: "/hello", node: &node{path: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{path: "/"}},
			{path: "/hello", node: &node{path: "/hello"}}}},
		{cap: 1, nodes: []lruNode{
			{path: "/", node: &node{path: "/"}},
			{path: "/hello", node: &node{path: "/hello"}},
			{path: "/world", node: &node{path: "/world"}}}},
		{cap: 2, nodes: []lruNode{
			{path: "/", node: &node{path: "/"}},
			{path: "/hello", node: &node{path: "/hello"}},
			{path: "/world", node: &node{path: "/world"}}}},
	}
)

func TestSplitPath(t *testing.T) {
	tcs := []struct {
		path  string
		parts []string
	}{
		{path: "/", parts: []string{"/"}},
		{path: "/hello", parts: []string{"/", "hello"}},
		{path: "/hello/", parts: []string{"/", "hello/"}},
		{path: "/hello/world", parts: []string{"/", "hello/", "world"}},
		{path: "/hello/world/", parts: []string{"/", "hello/", "world/"}},
		{path: "/hello/world/", parts: []string{"/", "hello/", "world/"}},
		{path: "/hello/{abc}/world", parts: []string{"/", "hello/", "{abc}/", "world"}},
		{path: "/hello/world/{abc}", parts: []string{"/", "hello/", "world/", "{abc}"}},
		{path: "/hello/world/{ab\\/c}/", parts: []string{"/", "hello/", "world/", "{ab\\/c}/"}},
		{
			path:  "/123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/world/{abc}-{def}/123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/",
			parts: []string{"/", "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world/", "world/", "{abc}-{def}/", "123-{abc:[a-z]{3-5}}-567-{haha:\\w+}--world++{a1:[0-9][0-9]?}ll/"},
		},
	}
	for _, tc := range tcs {
		parts := splitPath(tc.path)
		assert.True(t, reflect.DeepEqual(parts, tc.parts))
	}
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
		l := newLru(tc.cap)
		for _, n := range tc.nodes {
			l.put(n.path, n.node)
		}
		size := tc.cap
		if len(tc.nodes) < tc.cap {
			size = len(tc.nodes)
		}
		assert.Equal(t, size, l.len())
		for i := 1; i <= size; i++ {
			assert.Equal(t, tc.nodes[len(tc.nodes)-i].node.path, l.nodes.Remove(l.nodes.Front()).(*lruNode).node.path)
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
			n := l.get(tc.nodes[i].path)
			assert.Equal(t, tc.nodes[i].path, n.path)
			assert.Equal(t, tc.nodes[i].path, l.nodes.Front().Value.(*lruNode).path)
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

func TestParseExps(t *testing.T) {
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
		assert.Equal(t, len(parseExps(tc.part)), tc.cnt)
		assert.Equal(t, parseExps(tc.part), tc.exps)
	}
}
func TestParsePrefix(t *testing.T) {
	tcs := []struct {
		p   string
		idx int
		eof bool
	}{
		{p: "/", idx: 0, eof: true},
		{p: "/123", idx: 3, eof: true},
		{p: "/123/", idx: 4, eof: true},
		{p: "123/", idx: 3, eof: true},
		{p: "/{abc}", idx: 0, eof: false},
		{p: "/{abc}/", idx: 0, eof: false},
		{p: "{abc}/", idx: 5, eof: true},
		{p: "/123{abc}", idx: 0, eof: false},
		{p: "/123{abc}/", idx: 0, eof: false},
		{p: "123{abc}/", idx: 8, eof: true},
		{p: "/{abc}/123", idx: 0, eof: false},
		{p: "/{abc}/123/", idx: 0, eof: false},
		{p: "{abc}/123", idx: 5, eof: false},
		{p: "{abc}/123/", idx: 5, eof: false},
		{p: "/123{abc}/123", idx: 0, eof: false},
		{p: "/123{abc}/123/", idx: 0, eof: false},
		{p: "123{abc}/123", idx: 8, eof: false},
		{p: "123{abc}/123/", idx: 8, eof: false},
		{p: "/123/{abc}", idx: 4, eof: false},
		{p: "/123/{abc}/", idx: 4, eof: false},
		{p: "123/{abc}", idx: 3, eof: false},
		{p: "123/{abc}/", idx: 3, eof: false},
		{p: "/123/123{abc}", idx: 4, eof: false},
		{p: "/123/123{abc}/", idx: 4, eof: false},
		{p: "123/123{abc}", idx: 3, eof: false},
		{p: "123/123{abc}/", idx: 3, eof: false},
	}
	for _, tc := range tcs {
		idx, eof := parsePrefix(tc.p)
		assert.Equal(t, tc.idx, idx)
		assert.Equal(t, tc.eof, eof)
	}
}

func TestNewNode(t *testing.T) {
	tcs := []struct {
		part    string
		expKeys []string
		exps    map[string]string
	}{
		{part: "/", exps: nil},
		{part: "abc/", exps: nil},
		{part: "{abc}/", exps: map[string]string{"abc": `\S*`}},
		{part: "{abc:\\w+}/", exps: map[string]string{"abc": `\w+`}},
		{part: "{abc:[a-z][0-9]?}/", exps: map[string]string{"abc": `[a-z][0-9]?`}},
		{part: "hello-{abc:[a-z]+}!/", exps: map[string]string{"abc": `[a-z]+`}},
		{part: "hello-{abc:[a-z]+}!=bonjour-{def:[\\d][\\w]*}/", exps: map[string]string{"abc": `[a-z]+`, "def": "[\\d][\\w]*"}},
	}
	for _, tc := range tcs {
		n := newNode(tc.part)
		assert.Empty(t, n.path)
		assert.Equal(t, tc.part, n.part)
		if tc.exps == nil {
			assert.Nil(t, n.exps)
		} else {
			assert.Equal(t, len(n.exps), len(tc.exps))
			for k, v := range tc.exps {
				assert.Equal(t, v, n.exps[k].String())
			}
		}
		assert.NotNil(t, n.children)
		assert.Nil(t, n.params)
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

func TestNewRouter(t *testing.T) {
	r := newRouter()
	assert.NotNil(t, r)
	assert.NotNil(t, r.trees)
}
