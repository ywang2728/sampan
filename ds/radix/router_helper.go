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
	pathSeparator = '/'
	regexBegin    = '{'
	regexEnd      = '}'
	wildcardStar  = '*'
	wildcardColon = ':'
	keySeparators = `{:*`
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

	keyIter struct {
		cursor int
		keys   []Key[string]
	}

	keySeparator struct {
		begin rune
		end   rune
		cnt   atomic.Int32
	}
)

func newKeySeparator(begin rune, end rune) (ks *keySeparator) {
	return &keySeparator{begin, end, atomic.Int32{}}
}
func (ks *keySeparator) reset() {
	ks.cnt.Store(0)
}
func (ks *keySeparator) isBegin(c rune) bool {
	return ks.begin == c
}
func (ks *keySeparator) open() (opened bool) {
	return 0 != ks.cnt.Add(1)
}
func (ks *keySeparator) close() (closed bool) {
	return 0 == ks.cnt.Add(-1)
}
func (ks *keySeparator) closed() bool {
	return 0 == ks.cnt.Load()
}
func (ks *keySeparator) isEnd(c rune) bool {
	return ks.end == c
}
func (ks *keySeparator) load() int32 {
	return ks.cnt.Load()
}

func newKeyIter(keys ...Key[string]) KeyIterator[string] {
	return &keyIter{
		cursor: -1,
		keys:   keys,
	}
}

func (ki *keyIter) hasNext() bool {
	return ki.cursor+1 < len(ki.keys)

}
func (ki *keyIter) Next() Key[string] {
	if ki.hasNext() {
		ki.cursor++
		return ki.keys[ki.cursor]
	}
	return nil
}

// Main logic for parse the raw path, parse as much as the Key type allowed chars.
func buildKeyIterFunc(k string) (ki KeyIterator[string]) {
	if k == "" {
		return
	}
	if !strings.ContainsAny(k, keySeparators) {
		return newKeyIter(&staticKey{k})
	}
	var ks *keySeparator
	for i, j := 0, 0; i < len(k); i++ {
		switch k[i] {
		case wildcardStar:
			ks = newKeySeparator(wildcardStar, pathSeparator)
		case wildcardColon:
			ks = newKeySeparator(wildcardColon, pathSeparator)
		case regexBegin:
			ks = newKeySeparator(regexBegin, pathSeparator)
		}
		ks.open()

	}

	return
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
		if i := strings.Index(part, string(regexBegin)); i == -1 {
			rk.patterns.Put(part, nil)
			part = ""
		} else if i == 0 {
			ks := newKeySeparator(regexBegin, regexEnd)
			for ; i < len(part); i++ {
				if ks.isBegin(rune(part[i])) {
					ks.open()
				} else if ks.isEnd(rune(part[i])) && ks.close() {
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
			if !ks.closed() {
				log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
			}
		} else {
			rk.patterns.Put(part[:i], nil)
			part = part[i:]
		}
	}
	return
}

func hello() {
	print("hello")
}
