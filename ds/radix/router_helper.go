package radix

import (
	"errors"
	"github.com/ywang2728/sampan/ds/linkedhashmap"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
)

const (
	PathSeparator = '/'
	RegexBegin    = '{'
	RegexEnd      = '}'
	WildcardStar  = '*'
	WildcardColon = ':'
)

type (
	staticKey struct {
		value string
	}

	wildcardKey struct {
		staticKey
		params map[string]string
	}

	regexKey struct {
		wildcardKey
		patterns *linkedhashmap.Map[string, *regexp.Regexp]
	}

	keyIterator struct {
		index int
		keys  []Key[string]
	}

	reDelim struct {
		cnt atomic.Int32
	}
)

// Regex expression delimiter
func newReDelim() (delim *reDelim) {
	return &reDelim{atomic.Int32{}}
}
func (rd *reDelim) reset() {
	rd.cnt.Store(0)
}
func (rd *reDelim) open() (opened bool) {
	return 0 != rd.cnt.Add(1)
}
func (rd *reDelim) close() (closed bool) {
	return 0 == rd.cnt.Add(-1)
}
func (rd *reDelim) closed() bool {
	return 0 == rd.cnt.Load()
}
func (rd *reDelim) load() int32 {
	return rd.cnt.Load()
}

func (ki *keyIterator) hasNext() bool {
	return ki.index < len(ki.keys)

}
func (ki *keyIterator) Next() Key[string] {
	if ki.hasNext() {
		key := ki.keys[ki.index]
		ki.index++
		return key
	}
	return nil
}

func (sk *staticKey) Match(k string) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], p *map[string]string) {
	i, ln, lp := 0, len(sk.value), len(k)
	m := ln
	if m > lp {
		m = lp
	}
	for ; i < m; i++ {
		if sk.value[i] != k[i] {
			break
		}
	}
	if i > 0 {
		c = buildKeyIterFunc(sk.value[:i])
	}
	if i < ln {
		tn = buildKeyIterFunc(sk.value[i:])
	}
	if i < lp {
		tp = buildKeyIterFunc(k[i:])
	}
	return
}

func (sk *staticKey) String() string {
	return sk.value
}

func (sk *staticKey) Value() string {
	return sk.value
}

func (wk *wildcardKey) Match(k string) (c *string, tn *string, tp *string, p *map[string]string) {
	return
}

func (wk *wildcardKey) String() string {
	return wk.value
}

func (wk *wildcardKey) Value() string {
	return wk.value
}

func (rk *regexKey) Match(k string) (c *string, tn *string, tp *string, p *map[string]string) {
	return
}

func (rk *regexKey) String() string {
	return rk.value
}

func (rk *regexKey) Value() string {
	return rk.value
}

func (rk *regexKey) parsePatterns(part string) (patterns *linkedhashmap.Map[string, *regexp.Regexp]) {
	patterns = linkedhashmap.New[string, *regexp.Regexp]()
	for len(part) != 0 {
		if i := strings.Index(part, string(RegexBegin)); i == -1 {
			rk.patterns.Put(part, nil)
			part = ""
		} else if i == 0 {
			delim := newReDelim()
			for ; i < len(part); i++ {
				if part[i] == RegexBegin {
					delim.open()
				} else if part[i] == RegexEnd && delim.close() {
					var before, after string
					if i == len(part)-1 {
						before, after = part, ""
					} else {
						before, after = part[:i+1], part[i+1:]
					}
					l := len(before)
					if l < 3 {
						log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+before))
					}
					rk.patterns.Put(before, regexp.MustCompile(before[1:l-1]))
					part = after
					break
				}
			}
			if !delim.closed() {
				log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
			}
		} else {
			rk.patterns.Put(part[:i], nil)
			part = part[i:]
		}
	}
	return
}

// Main logic for parse the raw path, parse as much as the Key type allowed chars.
func buildKeyIterFunc(k string) (ki KeyIterator[string]) {

	ki = &keyIterator{
		keys: []Key[string]{&staticKey{k}},
	}
	return
}

func hello() {
	print("hello")
}
