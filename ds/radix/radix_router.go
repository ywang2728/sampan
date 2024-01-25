package radix

import (
	"errors"
	"fmt"
	"github.com/dlclark/regexp2"
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
)

type (
	keySeparator struct {
		bs  string
		es  string
		cnt atomic.Int32
	}
	keyIter struct {
		cursor int
		keys   []Key[string]
	}
	staticKey struct {
		value string
	}
	wildcardStarKey struct {
		value  string
		params map[string]string
	}
	wildcardColonKey struct {
		value  string
		params map[string]string
	}
	regexKey struct {
		value   string
		pattern *regexp.Regexp
		params  map[string]string
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
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\s)(?<target>\\s)`, 0):                   `[ \f\n\r\t\v]`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\S)(?<target>\\S)`, 0):                   `[^ \f\n\r\t\v]`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\x09)(?<target>\\x09)`, 0):               `\t`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\cI)(?<target>\\cI)`, 0):                 `\t`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\x0b)(?<target>\\x0b)`, 0):               `\v`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\cK)(?<target>\\cK)`, 0):                 `\v`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\w)(?<target>\\w)`, 0):                   `[A-Za-z0-9_]`,
		regexp2.MustCompile(`(?<prefix>\\*)(?=\\W)(?<target>\\W)`, 0):                   `[^A-Za-z0-9_]`,
	}
)

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

// KeyIter
// newKeyIter Function for parse the raw path, parse as much as the Key type allowed chars.
func newKeyIter(key string) KeyIterator[string] {
	var keys []Key[string]
	if strings.TrimSpace(key) != "" {
		var ks *keySeparator
		var kb int
		for cursor, ps := 0, -1; cursor < len(key); {
			switch string(key[cursor]) {
			case pathSeparator:
				ps = cursor
				cursor++
			case wildcardStar:
				if cursor+1 != len(key) {
					log.Fatalf("Key parsing error, %v", errors.New(fmt.Sprintf("Invalid wildcard key: %s at index: %d.", key, cursor)))
				}
				if kb < cursor {
					keys = append(keys, &staticKey{key[kb:cursor]})
				}
				keys = append(keys, &wildcardStarKey{value: wildcardStar, params: map[string]string{wildcardStar: ``}})
				cursor++
				kb = cursor
			case wildcardColon:
				if cursor != ps+1 {
					log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard key: %s at index: %d.`, key, cursor)))
				}
				if kb < cursor {
					keys = append(keys, &staticKey{key[kb:cursor]})
				}
				ks = newKeySeparator(wildcardColon, pathSeparator)
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
				if ks.closed() && len(part) > 1 && strings.HasPrefix(part, wildcardColon) && strings.Count(part, wildcardColon) == 1 && !strings.Contains(part, " ") {
					keys = append(keys, &wildcardColonKey{value: part, params: map[string]string{part[1:]: ""}})
					kb = cursor
				} else {
					log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid wildcard key: %s at index: %d.`, key, cursor)))
				}
			case regexBegin:
				if kb < cursor {
					keys = append(keys, &staticKey{key[kb:cursor]})
				}
				kb = cursor
				ks = newKeySeparator(regexBegin, regexEnd)
				ks.open()
				var part string
				for cursor++; ks.opened() && cursor < len(key); cursor++ {
					if _, ok := ks.openWith(string(key[cursor])); ok {
						continue
					} else if _, ok := ks.closeWith(string(key[cursor])); ok && ks.closed() {
						part = key[kb : cursor+1]
						break
					}
				}
				if cursor == len(key) && ks.opened() {
					part = key[kb:]
					ks.close()
				}
				if ks.closed() && len(part) > 3 && strings.HasPrefix(part, regexBegin) && strings.HasSuffix(part, regexEnd) {
					params := map[string]string{}
					compiled, err := regexp.Compile(part[1 : len(part)-1])
					if err == nil {
						for _, subExpName := range compiled.SubexpNames() {
							if subExpName != "" {
								params[subExpName] = ""
							}
						}
					} else {
						log.Fatalf("Key parsing error, #%v", errors.New(fmt.Sprintf(`Invalid regex key: %s at index: %d.`, part, cursor)))
					}
					keys = append(keys, &regexKey{value: part, pattern: compiled, params: params})
					kb = cursor + 1
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

func (ki *keyIter) String() string {
	var sb strings.Builder
	sb.WriteString("keyIter{")
	for _, k := range ki.keys {
		sb.WriteString(fmt.Sprintf("%v, ", k))
	}
	return strings.TrimRight(sb.String(), ", ") + "}"
}

// staticKey
func (sk *staticKey) String() string {
	return sk.value
}
func (sk *staticKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], override bool) {
	if k := ki.Next(); k != nil {
		if instKey, ok := k.(*staticKey); ok {
			i, ln, lp := 0, len(sk.value), len(instKey.value)
			m := ln
			if m > lp {
				m = lp
			}
			for ; i < m; i++ {
				if sk.value[i] != instKey.value[i] {
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
						if instKey2, ok := instKI.keys[instKI.cursor+1].(*staticKey); ok {
							instKey2.value = instKey.value[i:] + instKey2.value
						} else {
							tpKeys = append(tpKeys, &staticKey{instKey.value[i:]})
						}
						tpKeys = append(tpKeys, instKI.keys[instKI.cursor+1:]...)
					} else if i == lp {
						tpKeys = append(tpKeys, instKI.keys[instKI.cursor+1:]...)
					}
				}
			} else if i < lp {
				tpKeys = append(tpKeys, &staticKey{instKey.value[i:]})
			}
			if len(tpKeys) > 0 {
				tp = &keyIter{-1, tpKeys}
			}
		} else if _, ok := k.(*wildcardStarKey); ok {
			c, override = &keyIter{-1, []Key[string]{k}}, true
		} else if _, ok := k.(*wildcardColonKey); ok {
			c, override = &keyIter{-1, []Key[string]{k}}, true
			if ki.HasNext() {
				if instKI, ok := ki.(*keyIter); ok {
					tp = &keyIter{-1, instKI.keys[instKI.cursor+1:]}
				}
			}
		} else {
			if override = k.(*regexKey).pattern.MatchString(sk.value); override {
				c = &keyIter{-1, []Key[string]{k}}
				if instKI, ok := ki.(*keyIter); ok && ki.HasNext() {
					tp = &keyIter{-1, instKI.keys[instKI.cursor+1:]}
				}
			} else {
				tn = &keyIter{-1, []Key[string]{sk}}
				if instKI, ok := ki.(*keyIter); ok {
					tp = &keyIter{-1, instKI.keys[instKI.cursor:]}
				}
			}
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
	return wsk.value
}
func (wsk *wildcardStarKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], override bool) {
	if ki.HasNext() {
		instKI := ki.(*keyIter)
		ki.(*keyIter).cursor = len(instKI.keys)
		c = &keyIter{-1, []Key[string]{wsk}}
	}
	return
}
func (wsk *wildcardStarKey) Match(s string) (t string, p map[string]string, matched bool) {
	p[wildcardStar], matched = s, true
	return
}

// wildcardColonKey
func (wck *wildcardColonKey) String() string {
	return wck.value
}
func (wck *wildcardColonKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], override bool) {
	if k := ki.Next(); k != nil {
		if instKey, ok := k.(*staticKey); ok {
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
		} else if _, ok := k.(*wildcardStarKey); ok {

		} else if _, ok := k.(*wildcardColonKey); ok {

		} else {
			//_ := k.(*regexKey)
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
	return rk.value
}
func (rk *regexKey) MatchIterator(ki KeyIterator[string]) (c KeyIterator[string], tn KeyIterator[string], tp KeyIterator[string], override bool) {
	if ki.HasNext() {
		instKI := ki.(*keyIter)
		i := instKI.cursor + 1
		if instKey, ok := instKI.keys[i].(*staticKey); ok {
			if loc := rk.pattern.FindStringIndex(instKey.value); loc != nil && loc[0] == 0 {
				toBeMatched := instKey.value[loc[0]:loc[1]]
				if tail, matched := strings.CutPrefix(instKey.value, toBeMatched); matched {
					if tail == "" {
						c = &keyIter{-1, []Key[string]{rk}}
					} else {
						c = &keyIter{-1, []Key[string]{&staticKey{toBeMatched}}}
						instKey.value = tail
						tp = &keyIter{-1, instKI.keys[i:]}
					}
				}
			}
		} else if _, ok := instKI.keys[i].(*wildcardStarKey); ok {
			c = &keyIter{-1, []Key[string]{rk}}
		} else if _, ok := instKI.keys[i].(*wildcardColonKey); ok {
			c = &keyIter{-1, []Key[string]{rk}}
		} else {
			instKey, _ := instKI.keys[i].(*regexKey)
			if formatRePattern(rk.value) == formatRePattern(instKey.value) {
				c = &keyIter{-1, []Key[string]{rk}}
				tp = &keyIter{-1, instKI.keys[i+1:]}
			} else {
				tn = &keyIter{-1, []Key[string]{rk}}
				tp = &keyIter{-1, instKI.keys[i:]}
			}
		}
	}
	return
}
func (rk *regexKey) Match(s string) (t string, p map[string]string, matched bool) {
	if loc := rk.pattern.FindStringIndex(s); loc != nil && loc[0] == 0 {
		toBeMatched := s[loc[0]:loc[1]]
		if s, matched = strings.CutPrefix(s, toBeMatched); matched {
			match := rk.pattern.FindStringSubmatch(s[loc[0]:loc[1]])
			for i, name := range rk.pattern.SubexpNames() {
				if i != 0 && name != "" {
					p[name] = match[i]
				}
			}
		}
	}
	t = s
	return
}
