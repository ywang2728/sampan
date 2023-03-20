package sampan

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

/*func TestLcp(t *testing.T) {
	cases := []struct {
		s1  string
		s2  string
		idx int
	}{
		{s1: "/", s2: "1", idx: -1},
		{s1: "/1", s2: "1", idx: -1},
		{s1: "/1/", s2: "/2/", idx: 0},
		{s1: "/", s2: "/", idx: 0},
		{s1: "/", s2: "/123", idx: 0},
		{s1: "/123", s2: "/", idx: 0},
		{s1: "/12/456", s2: "/12/567", idx: 3},
		{s1: "/12/456", s2: "/12/456/", idx: 3},
		{s1: "/:123/456", s2: "/*123/567", idx: 0},
		{s1: "/123/4/56", s2: "/123/4/57", idx: 6},
		{s1: "/123/4/5/6", s2: "/123/4/5/7", idx: 8},
		{s1: "/123/4/56/", s2: "/123/4/56", idx: 6},
	}
	asst := assert.New(t)

	for _, tc := range cases {
		asst.Equal(tc.idx, lcp(&tc.s1, &tc.s2))
	}
}*/

func TestNewRadix(t *testing.T) {
	r := newRadix()
	assert.NotNil(t, r)
}

func TestNewRouter(t *testing.T) {
	r := newRouter()
	assert.NotNil(t, r)
	assert.NotNil(t, r.trees)
	assert.NotNil(t, r.handlers)
}

func TestParseExps(t *testing.T) {
	cases := []struct {
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
	for _, tc := range cases {
		assert.Equal(t, len(parseExps(tc.part)), tc.cnt)
		assert.Equal(t, parseExps(tc.part), tc.exps)
	}
}

func TestNewNode(t *testing.T) {
	cases := []struct {
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
	for _, tc := range cases {
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
