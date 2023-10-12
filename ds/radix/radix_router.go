package radix

import (
	"errors"
	"fmt"
	"github.com/ywang2728/sampan/ds/linkedhashmap"
	"log"
	"regexp"
	"strings"
	"sync/atomic"
)

const (
	pathSeparator = `/`
	regexBegin    = `{`
	regexEnd      = `}`
	wildcardStar  = `*`
	wildcardColon = `:`
	keySeparators = `{:*`
)

type (
	staticKey struct {
		value string
	}

	wildcardStarKey struct {
		value  string
		prefix string
		suffix string
		params map[string]string
	}

	wildcardColonKey struct {
		value  string
		params map[string]string
	}

	regexKey struct {
		value    []string
		patterns *linkedhashmap.Map[string, *regexp.Regexp]
		params   map[string]string
	}

	keyIter struct {
		cursor int
		keys   []Key[string]
	}

	keySeparator struct {
		bs  string
		es  string
		cnt atomic.Int32
	}
)

func newKeySeparator(begin string, end string) (ks *keySeparator) {
	return &keySeparator{bs: begin, es: end, cnt: atomic.Int32{}}
}
func (ks *keySeparator) reset() {
	ks.cnt.Store(0)
}
func (ks *keySeparator) isOpened() bool {
	return 0 != ks.cnt.Load()
}
func (ks *keySeparator) open() (times int, opened bool) {
	before := ks.cnt.Load()
	opened = before < ks.cnt.Add(1)
	times = int(ks.cnt.Load())
	return
}
func (ks *keySeparator) openWith(s string) (times int, opened bool) {
	before := ks.cnt.Load()
	if ks.bs == s {
		opened = before < ks.cnt.Add(1)
	}
	times = int(ks.cnt.Load())
	return
}
func (ks *keySeparator) closeWith(s string) (times int, closed bool) {
	before := ks.cnt.Load()
	if ks.es == s {
		closed = before > ks.cnt.Add(-1)
	}
	times = int(ks.cnt.Load())
	return
}

func (ks *keySeparator) close() (times int, closed bool) {
	before := ks.cnt.Load()
	after := ks.cnt.Add(-1)
	closed = before > after && after >= 0
	times = int(ks.cnt.Load())
	return
}
func (ks *keySeparator) isClosed() bool {
	return 0 == ks.cnt.Load()
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
	return sk.Value()
}

func (sk *staticKey) Value() string {
	return sk.value
}

func (wk *wildcardStarKey) Match(k string) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], p *map[string]string) {
	fmt.Println(k)
	return
}

func (wk *wildcardStarKey) String() string {
	return wk.Value()
}

func (wk *wildcardStarKey) Value() string {
	return wk.value
}

func (wk *wildcardColonKey) Match(k string) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], p *map[string]string) {
	fmt.Println(k)
	return
}

func (wk *wildcardColonKey) String() string {
	return wk.Value()
}

func (wk *wildcardColonKey) Value() string {
	return wk.value
}

func (rk *regexKey) Match(k string) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], p *map[string]string) {
	fmt.Println(k)
	return
}

func (rk *regexKey) String() string {
	return rk.Value()
}

func (rk *regexKey) Value() string {
	return fmt.Sprint(rk.value)
}

func (rk *regexKey) parsePatterns(part string) (patterns *linkedhashmap.Map[string, *regexp.Regexp]) {
	patterns = linkedhashmap.New[string, *regexp.Regexp]()
	for len(part) != 0 {
		if i := strings.Index(part, regexBegin); i == -1 {
			rk.patterns.Put(part, nil)
			part = ""
		} else if i == 0 {
			ks := newKeySeparator(regexBegin, regexEnd)
			for ; i < len(part); i++ {
				if _, ok := ks.openWith(string(part[i])); ok {
					continue
				} else if _, ok := ks.closeWith(string(part[i])); ok && ks.isClosed() {
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
			if !ks.isClosed() {
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
func buildKeyIterFunc(key string) (ki KeyIterator[string]) {
	if strings.TrimSpace(key) == "" {
		return newKeyIter()
	}
	if !strings.ContainsAny(key, keySeparators) {
		return newKeyIter(&staticKey{key})
	}
	var keys []Key[string]
	var ks *keySeparator
	for cursor, i, ps := 0, 0, -1; cursor < len(key); {
		switch string(key[cursor]) {
		case pathSeparator:
			ps = cursor
			cursor++
		case wildcardStar:
			if i <= ps {
				keys = append(keys, &staticKey{key[i : ps+1]})
			}
			ks = newKeySeparator(wildcardStar, pathSeparator)
			ks.open()
			var part string
			for cursor++; ks.isOpened() && cursor < len(key); cursor++ {
				if _, ok := ks.openWith(string(key[cursor])); ok {
					continue
				} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.isClosed() {
					part = key[ps+1 : cursor]
					break
				}
			}
			if cursor == len(key) && ks.isOpened() {
				part = key[ps+1:]
				ks.close()
			}
			if ks.isClosed() {
				if strings.Count(part, wildcardStar) != 1 || strings.Contains(part, " ") {
					log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard star key: %s at index: %d.`, key, cursor)))
				}
				pref, suf, _ := strings.Cut(part, wildcardStar)
				keys = append(keys, &wildcardStarKey{value: part, prefix: pref, suffix: suf, params: map[string]string{part: ""}})
			} else {
				log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard star key: %s at index: %d.`, key, cursor)))
			}
		case wildcardColon:
			if i <= ps {
				keys = append(keys, &staticKey{key[i : ps+1]})
			}
			var part string
			for ks = newKeySeparator(wildcardColon, pathSeparator); ks.isOpened() && cursor < len(key); cursor++ {
				if _, ok := ks.openWith(string(key[cursor])); ok {
					continue
				} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.isClosed() {
					part = key[ps+1 : cursor]
				} else if cursor == len(key)-1 {
					part = key[ps+1:]
					ks.close()
				}
			}
			if ks.isClosed() {
				if len(part) == 1 || !strings.HasPrefix(part, wildcardColon) || strings.Count(part, wildcardColon) != 1 || strings.Contains(part, " ") {
					log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard colon key: %s at index: %d.`, key, cursor)))
				}
				keys = append(keys, &wildcardColonKey{value: part, params: map[string]string{part[cursor+1:]: ""}})
			} else {
				log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard colon key: %s at index: %d.`, key, cursor)))
			}
		case regexBegin:
			if i <= ps {
				keys = append(keys, &staticKey{key[i : ps+1]})
			}
			var parts []string
			var reBgn, reEnd int
			patterns := linkedhashmap.New[string, *regexp.Regexp]()
			if ps != -1 {
				reEnd = ps
			} else {
				reEnd = cursor
			}
			for ks = newKeySeparator(regexBegin, regexEnd); cursor < len(key) && (ks.isOpened() || pathSeparator != string(key[cursor])); cursor++ {
				if times, ok := ks.openWith(string(key[cursor])); ok {
					if times == 1 {
						reBgn = cursor
						if reEnd+1 < reBgn {
							part := key[reEnd+1 : reBgn]
							if strings.Contains(part, " ") {
								log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid regex key: %s at index: %d.`, key, cursor)))
							}
							parts = append(parts, part)
						}
					}
				} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.isClosed() {
					reEnd = cursor
					if reBgn < reEnd {
						part := key[reBgn : reEnd+1]
						if !strings.HasPrefix(part, regexBegin) || !strings.HasSuffix(part, regexEnd) {
							log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid regex key: %s at index: %d.`, key, cursor)))
						}
						parts = append(parts)
						patterns.Put(part, regexp.MustCompile(part))
					}
				}
			}
			if reEnd+1 < cursor {
				part := key[reEnd+1 : cursor]
				if strings.Contains(part, " ") {
					log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid regex key: %s at index: %d.`, key, cursor)))
				}
				parts = append(parts, key[reEnd+1:cursor])
			}
			if ks.isClosed() && len(parts) > 0 {
				keys = append(keys, &regexKey{value: parts, params: map[string]string{}, patterns: patterns})
			} else {
				log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid Regex key: %s at index: %d.`, key, cursor)))
			}
		default:
			cursor++
		}
	}
	return newKeyIter(keys...)
}

func hello() {
	print("hello")
}
