package radix

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
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

	keySeparator struct {
		bs  string
		es  string
		cnt atomic.Int32
	}
	keyIter struct {
		cursor int
		keys   []Key[string]
	}
)

var (
	reFormatPatterns = map[*regexp2.Regexp]string{
		regexp2.MustCompile(`(?<prefix>\\*)(?=\(\?P<[^>]*>)(?<target>\(\?P<[^>]*>)`, 0): `(`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\d)(?<target>\\d)`, 0):                   `[0-9]`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\D)(?<target>\\D)`, 0):                   `[^0-9]`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\x0c)(?<target>\\x0c)`, 0):               `\f`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\cL)(?<target>\\cL)`, 0):                 `\f`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\x0a)(?<target>\\x0a)`, 0):               `\n`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\cJ)(?<target>\\cJ)`, 0):                 `\n`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\x0d)(?<target>\\x0d)`, 0):               `\r`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\cM)(?<target>\\cM)`, 0):                 `\r`,
	}
)

// staticKey
func (sk *staticKey) String() string {
	return sk.value
}
func (sk *staticKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string]) {
	if ki.HasNext() {
		if kk, ok := ki.Next().(*staticKey); ok {
			i, ln, lp := 0, len(sk.value), len(kk.value)
			m := ln
			if m > lp {
				m = lp
			}
			for ; i < m; i++ {
				if sk.value[i] != kk.value[i] {
					break
				}
			}
			if i > 0 {
				c = newKeyIter(sk.value[:i])
			}
			if i < ln {
				tn = newKeyIter(sk.value[i:])
			}
			var tpKeys []Key[string]
			if ki.HasNext() {
				if instKI, ok := ki.(*keyIter); ok {
					if i == 0 {
						tpKeys = append(tpKeys, instKI.keys[instKI.cursor:]...)
					} else if i < lp {
						if instKey, ok := instKI.keys[instKI.cursor+1].(*staticKey); ok {
							instKey.value = kk.value[i:] + instKey.value
						} else {
							tpKeys = append(tpKeys, &staticKey{kk.value[i:]})
						}
						tpKeys = append(tpKeys, instKI.keys[instKI.cursor+1:]...)
					} else if i == lp {
						tpKeys = append(tpKeys, instKI.keys[instKI.cursor+1:]...)
					}
				}
			} else if i < lp {
				tpKeys = append(tpKeys, &staticKey{kk.value[i:]})
			}
			if len(tpKeys) > 0 {
				tp = &keyIter{-1, tpKeys}
			}
		} else {
			tp = ki
		}
	}
	return
}
func (sk *staticKey) Match(s string) (t string, p map[string]string, matched bool) {
	t, matched = strings.CutPrefix(s, sk.value)
	return
}

// wildcardStarKey
func (wsk *wildcardStarKey) String() string {
	return wsk.prefix + wsk.value + wsk.suffix
}
func (wsk *wildcardStarKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string]) {
	if ki.HasNext() {
		instKI := ki.(*keyIter)
		sb := strings.Builder{}
		matched := false
		for i, l := instKI.cursor+1, len(instKI.keys); i < l && !matched; i++ {
			sb.WriteString(instKI.keys[i].String())
			if _, ok := strings.CutPrefix(sb.String(), wsk.prefix); ok {
				if wsk.suffix == "" {
					ki.(*keyIter).cursor = l
					c = ki
					matched = true
				} else if _, ok := strings.CutSuffix(sb.String(), wsk.suffix); ok {
					ki.(*keyIter).cursor = i
					c = ki
					if i < l-1 {
						tp = &keyIter{i + 1, instKI.keys}
					}
					matched = true
				}
			}
		}
		if !matched {
			tp = ki
		}
	}
	return
}
func (wsk *wildcardStarKey) Match(s string) (t string, p map[string]string, matched bool) {
	if t, matched = strings.CutPrefix(s, wsk.prefix); matched {
		if i := strings.Index(s, wsk.suffix); i > -1 {
			p[wildcardStar], t, matched = strings.Cut(t, wsk.suffix)
		} else {
			t = s
		}
	} else {
		t = s
	}
	return
}

// wildcardColonKey
func (wck *wildcardColonKey) String() string {
	return wck.value
}
func (wck *wildcardColonKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string]) {
	if ki.HasNext() {
		instKI := ki.(*keyIter)
		i := instKI.cursor + 1
		if instKey, ok := instKI.keys[i].(*staticKey); ok {
			j := strings.Index(instKey.value, pathSeparator)
			if j > 0 {
				c = &keyIter{cursor: -1, keys: []Key[string]{&staticKey{instKey.value[:j]}}}
				if j < len(instKey.value) {
					instKey.value = instKey.value[j:]
					tp = ki
				}
			} else {
				tp = ki
			}
		}
	}
	return
}
func (wck *wildcardColonKey) Match(s string) (t string, p map[string]string, matched bool) {
	if i := strings.Index(s, pathSeparator); i > 0 {
		p[wck.value], t, matched = s[:i], s[i:], true
	} else {
		t = s
	}
	return
}

// regexKey
func (rk *regexKey) String() string {
	return strings.Join(rk.value, "")
}
func (rk *regexKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string]) {
	//if kk, ok := k.(*staticKey); ok {
	//	marched := true
	//	for i, kkv := 0, kk.value; i < len(rk.value) && marched; i++ {
	//		if compiled, ok := rk.patterns.Get(rk.value[i]); ok {
	//			marched = compiled.MatchString(kkv)
	//		} else {
	//			kkv, marched = strings.CutPrefix(kkv, rk.value[i])
	//		}
	//	}
	//	if marched {
	//		c = newKeyIter(rk)
	//		//TODO parse regex group values to Params
	//	}
	//} else {
	//	tp = newKeyIter(k)
	//}
	return
}
func (rk *regexKey) Match(s string) (t string, p map[string]string, matched bool) {
	matched = true
	sb := strings.Builder{}
	for _, raw := range rk.value {
		if compiled, ok := rk.patterns.Get(raw); ok {
			if loc := compiled.FindStringIndex(s); loc != nil && loc[0] == 0 {
				if loc[1] == len(s) || s[loc[1]] == '/' {
					toBeMatched := s[loc[0]:loc[1]]
					if s, matched = strings.CutPrefix(s, toBeMatched); matched {
						sb.WriteString(toBeMatched)
						if names := compiled.SubexpNames(); len(names) > 1 {
							for i, match := range compiled.FindStringSubmatch(toBeMatched) {
								if len(names[i]) != 0 {
									p[names[i]] = match
								}
							}
						}
					}
				} else {
					matched = false
				}
			} else {
				matched = false
			}
		} else {
			if s, matched = strings.CutPrefix(s, raw); matched {
				sb.WriteString(raw)
			} else {
				matched = false
			}
		}
		if matched {
			t = s
		}
	}
	return
}

// KeySeparator
func newKeySeparator(begin string, end string) (ks *keySeparator) {
	return &keySeparator{bs: begin, es: end, cnt: atomic.Int32{}}
}
func (ks *keySeparator) reset() {
	ks.cnt.Store(0)
}
func (ks *keySeparator) opened() bool {
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
func (ks *keySeparator) closed() bool {
	return 0 == ks.cnt.Load()
}

// Function for sanitize regex pattern and remove group name for regex pattern match.
func formatRePattern(p string) string {
	for k, v := range reFormatPatterns {
		p, _ = k.ReplaceFunc(p, func(m regexp2.Match) string {
			pg := m.GroupByName("prefix")
			if pg.Length&1 == 0 {
				return pg.String() + v
			} else {
				return pg.String() + m.GroupByName("target").String()
			}
		}, -1, -1)
	}
	return p
}

// KeyIter
// newKeyIter Function for parse the raw path, parse as much as the Key type allowed chars.
func newKeyIter(key string) KeyIterator[string] {
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
					for cursor++; ks.opened() && cursor < len(key); cursor++ {
						if _, ok := ks.openWith(string(key[cursor])); ok {
							continue
						} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.closed() {
							part = key[ps+1 : cursor]
							break
						}
					}
					if cursor == len(key) && ks.opened() {
						part = key[ps+1:]
						ks.close()
					}
					if ks.closed() {
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
					for cursor++; cursor < len(key) && (ks.opened() || pathSeparator != string(key[cursor])); cursor++ {
						if times, ok := ks.openWith(string(key[cursor])); ok {
							if times == 1 {
								reBgn = cursor
								if reEnd+1 < reBgn {
									parts = append(parts, key[reEnd+1:reBgn])
								}
							}
						} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.closed() {
							reEnd = cursor
							if reBgn < reEnd {
								part := key[reBgn : reEnd+1]
								compiled, err := regexp.Compile(part)
								if err == nil {
									parts = append(parts, part)
									patterns.Put(part, compiled)
									for _, subExpName := range compiled.SubexpNames() {
										if subExpName != "" {
											params[subExpName] = ""
										}
									}
								} else {
									log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid regex key: %s at index: %d.`, part, reBgn)))
								}
							}
						}
					}
					if reEnd+1 < cursor {
						parts = append(parts, key[reEnd+1:cursor])
					}
					if ks.closed() && len(parts) > 0 {
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
	return &keyIter{-1, keys}
}

func (ki *keyIter) Reset() {
	ki.cursor = -1
}
func (ki *keyIter) HasNext() bool {
	return ki.cursor+1 < len(ki.keys)
}
func (ki *keyIter) Next() Key[string] {
	if ki.HasNext() {
		ki.cursor++
		return ki.keys[ki.cursor]
	}
	return nil
}
