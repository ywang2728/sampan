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
func buildKeyIterFunc(key string) KeyIterator[string] {
	var keys []Key[string]
	if strings.TrimSpace(key) != "" {
		if strings.ContainsAny(key, keySeparators) {
			var ks *keySeparator
			var kb int
			for cursor, ps := 0, -1; cursor < len(key); {
				switch string(key[cursor]) {
				case pathSeparator:
					ps = cursor
					cursor++
				case wildcardStar, wildcardColon:
					if kb <= ps {
						keys = append(keys, &staticKey{key[kb : ps+1]})
					}
					wildcard := string(key[cursor])
					ks = newKeySeparator(wildcard, pathSeparator)
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
						kb = cursor
						if strings.Count(part, wildcard) != 1 || strings.Contains(part, " ") {
							log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard key: %s at index: %d.`, key, cursor)))
						}
						if wildcard == wildcardStar {
							pref, suf, _ := strings.Cut(part, wildcard)
							keys = append(keys, &wildcardStarKey{value: wildcard, prefix: pref, suffix: suf, params: map[string]string{wildcard: ""}})
						} else {
							if len(part) == 1 || !strings.HasPrefix(part, wildcard) {
								log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard key: %s at index: %d.`, key, cursor)))
							}
							suf, _ := strings.CutPrefix(part, wildcard)
							keys = append(keys, &wildcardColonKey{value: part, params: map[string]string{suf: ""}})
						}
					} else {
						log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard key: %s at index: %d.`, key, cursor)))
					}
				case regexBegin:
					if kb <= ps {
						keys = append(keys, &staticKey{key[kb : ps+1]})
					}
					var parts []string
					if ps+1 < cursor {
						parts = append(parts, key[ps+1:cursor])
					}
					patterns := linkedhashmap.New[string, *regexp.Regexp]()
					params := map[string]string{}
					ks = newKeySeparator(regexBegin, regexEnd)
					ks.open()
					reBgn, reEnd := cursor, 0
					for cursor++; cursor < len(key) && (ks.isOpened() || pathSeparator != string(key[cursor])); cursor++ {
						if times, ok := ks.openWith(string(key[cursor])); ok {
							if times == 1 {
								reBgn = cursor
								if reEnd+1 < reBgn {
									parts = append(parts, key[reEnd+1:reBgn])
								}
							}
						} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.isClosed() {
							reEnd = cursor
							if reBgn < reEnd {
								part := key[reBgn : reEnd+1]
								parts = append(parts, part)
								compiled := regexp.MustCompile(part)
								patterns.Put(part, compiled)
								for _, subExpName := range compiled.SubexpNames() {
									if subExpName != "" {
										params[subExpName] = ""
									}
								}
							}
						}
					}
					if reEnd+1 < cursor {
						parts = append(parts, key[reEnd+1:cursor])
					}
					if ks.isClosed() && len(parts) > 0 {
						kb = cursor
						keys = append(keys, &regexKey{value: parts, params: params, patterns: patterns})
					} else {
						log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid Regex key: %s at index: %d.`, key, cursor)))
					}
				default:
					cursor++
				}
			}
			if kb < len(key) {
				keys = append(keys, &staticKey{key[kb:]})
			}
		} else {
			keys = append(keys, &staticKey{key})
		}
	}
	return newKeyIter(keys...)
}

func hello() {
	print("hello")
}
