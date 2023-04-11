package sampan

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	LruCapacity = 255
	ReDelimBgn  = '{'
	ReDelimMid  = ':'
	ReDelimEnd  = '}'
)

var (
	ReKeyPattern = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
)

type (
	reDelim struct {
		cnt atomic.Int32
	}

	reMap struct {
		divs []string
		exps map[string]*regexp.Regexp
	}

	node struct {
		part       string
		rePatterns *reMap
		handler    func(*Context)
		//store non-regex nodes by first segment of tail of path as map key.
		children map[string]*node
		//store regex nodes, keep the insert order for seeking.
		reChildren []*node
	}

	lruNode struct {
		path   string
		node   *node
		params map[string]string
	}

	lru struct {
		nodes *list.List
		paths map[string]*list.Element
		cap   int
		mutex sync.Mutex
	}

	radix struct {
		cache *lru
		root  *node
		size  int
		mutex sync.RWMutex
	}

	router struct {
		trees map[string]*radix
	}
)

// LRU cache for storing the latest recent URL
func newLru(cap int) *lru {
	return &lru{
		cap:   cap,
		nodes: list.New(),
		paths: make(map[string]*list.Element, cap),
	}
}
func (l *lru) clear() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.nodes.Init()
}
func (l *lru) len() int {
	return l.nodes.Len()
}
func (l *lru) put(path string, node *node, params map[string]string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if e, ok := l.paths[path]; ok {
		l.nodes.MoveToFront(e)
	} else {
		if l.nodes.Len() == l.cap {
			delete(l.paths, l.nodes.Remove(l.nodes.Back()).(*lruNode).path)
		}
		l.paths[path] = l.nodes.PushFront(&lruNode{path: path, node: node, params: params})
	}
}
func (l *lru) get(path string) (n *node, params map[string]string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if e, ok := l.paths[path]; ok {
		l.nodes.MoveToFront(e)
		n = e.Value.(*lruNode).node
		params = e.Value.(*lruNode).params
	}
	return
}
func (l *lru) delete(path string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if e, ok := l.paths[path]; ok {
		delete(l.paths, l.nodes.Remove(e).(*lruNode).path)
	}
}

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

// parse the prefix from path for one node base on the difference of regex segment and plain text segment.
func parsePref(path string) (idx int, eof bool) {
	if lp, i, j := len(path), strings.Index(path, string(ReDelimBgn)), strings.Index(path, "/"); i == -1 || j == -1 {
		idx, eof = lp-1, true
	} else if i < j {
		for delim := newReDelim(); i < lp; i++ {
			if path[i] == ReDelimBgn {
				delim.open()
			} else if path[i] == ReDelimEnd {
				delim.close()
			} else if path[i] == '/' && delim.closed() {
				idx = i
				break
			}
		}
		if i == lp || i == lp-1 {
			idx = lp - 1
			eof = true
		}
	} else {
		idx = strings.LastIndex(path[:i], "/")
	}
	return
}

// Parse first segment of tail path as key.
func parseKey(path string) (key string) {
	i := 0
	for delim := newReDelim(); i < len(path); i++ {
		if path[i] == ReDelimBgn {
			delim.open()
		} else if path[i] == ReDelimEnd {
			delim.close()
		} else if path[i] == '/' && delim.closed() {
			key = path[:i+1]
			break
		}
	}
	if i == len(path) {
		key = path
	}
	return
}

func parseRePattern(part string) (bef string, aft string, isRe bool) {
	if i := strings.Index(part, string(ReDelimBgn)); i == -1 {
		bef, aft, isRe = part, "", false
	} else if i == 0 {
		delim := newReDelim()
		for ; i < len(part); i++ {
			if part[i] == ReDelimBgn {
				delim.open()
			} else if part[i] == ReDelimEnd {
				if delim.close() {
					if i == len(part)-1 {
						bef, aft, isRe = part, "", true
					} else {
						bef, aft, isRe = part[:i+1], part[i+1:], true
					}
					bared := bef[1 : len(bef)-1]
					if strings.Contains(bared, string(ReDelimMid)) {
						if strings.HasPrefix(bared, string(ReDelimMid)) || strings.HasSuffix(bared, string(ReDelimMid)) ||
							!ReKeyPattern.MatchString(bared[:strings.Index(bared, string(ReDelimMid))]) {
							log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+bef))
						}
					} else {
						if !ReKeyPattern.MatchString(bared) {
							log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+bef))
						}
					}
					break
				}
			}
		}
		if !delim.closed() {
			log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
		}
	} else {
		bef, aft, isRe = part[:i], part[i:], false
	}
	return
}

func parseRe(part string) (matches []string) {
	expKeyRe := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	var sb strings.Builder
	delim := newReDelim()
	for i := strings.Index(part, string(ReDelimBgn)); i < len(part); i++ {
		if part[i] == ReDelimBgn {
			delim.open()
			if delim.load() != 1 {
				sb.WriteString(string(part[i]))
			}
		} else if part[i] == ReDelimEnd {
			delim.close()
			if delim.load() != 0 {
				sb.WriteString(string(part[i]))
			}
			if delim.closed() {
				s := sb.String()
				if strings.Contains(s, string(ReDelimMid)) {
					if strings.HasPrefix(s, string(ReDelimMid)) || strings.HasSuffix(s, string(ReDelimMid)) ||
						!expKeyRe.MatchString(s[0:strings.Index(s, string(ReDelimMid))]) {
						log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+s))
					}
				} else {
					if !expKeyRe.MatchString(s) {
						log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+s))
					}
				}
				matches = append(matches, s)
				if strings.Contains(part[i:], string(ReDelimBgn)) {
					sb.Reset()
				} else {
					break
				}
			}
		} else if !delim.closed() {
			sb.WriteString(string(part[i]))
		}
	}
	if !delim.closed() {
		log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
	}
	return
}

// regex map to handle the regex expression with inserting order
func newReMap(part string) (m *reMap) {
	m = &reMap{
		divs: []string{},
		exps: make(map[string]*regexp.Regexp),
	}
	var (
		seg  string
		tail = part
		isRe bool
	)
	for {
		seg, tail, isRe = parseRePattern(tail)
		if isRe {
			k, v, ok := strings.Cut(seg[1:len(seg)-1], string(ReDelimMid))
			if _, exist := m.exps[k]; exist {
				log.Fatalf("Expression parsing error, #%v", errors.New(`duplicated expression key:`+k))
			}
			if ok {
				m.exps[k] = regexp.MustCompile(v)
			} else {
				m.exps[k] = regexp.MustCompile(`\S*`)
			}
		}
		m.divs = append(m.divs, seg)
		if len(tail) == 0 {
			break
		}
	}
	return
}
func (rm *reMap) len() (l int) {
	return len(rm.divs)
}

func newNode(part string) (n *node) {
	n = &node{
		children:   make(map[string]*node),
		reChildren: make([]*node, 0),
	}
	if len(part) > 0 {
		n.part = part
		if strings.Contains(part, string(ReDelimBgn)) {
			n.rePatterns = newReMap(part)
		}
	}
	return
}

func (n *node) getReChild(part string) (child *node, ok bool) {
	for _, c := range n.reChildren {
		if c.part == part {
			return c, true
		}
	}
	return
}

// find the longest common prefix from the index position by "/", wildcard will be considered as different part and be treated as single node.
func commonPrefix(s1, s2 string) (p string, s int) {
	l1, l2 := len(s1), len(s2)
	min := l1
	if min > l2 {
		min = l2
	}
	s = -1
	i := 0
	for i < min {
		if s1[i] != s2[i] {
			break
		}
		if s1[i] == '/' {
			s = i
		}
		i++
	}
	if i == min {
		p = s1[0:min]
	} else {
		p = s1[0 : s+1]
	}
	return
}

func newRadix() *radix {
	return &radix{
		cache: newLru(LruCapacity),
	}
}

func (r *radix) clear() {
	if r != nil {
		r.mutex.Lock()
		defer r.mutex.Unlock()
		r.root = nil
		r.cache.clear()
		r.size = 0
	}
}

func (r *radix) len() int {
	if r == nil {
		return 0
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.root == nil {
		return 0
	}
	return r.size
}

func (r *radix) stringRec(n *node, l int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat("#", l))
	output.WriteString(fmt.Sprintf(" %p : %+v\n", n, n))
	for _, child := range n.children {
		output.WriteString(r.stringRec(child, l+1))
	}
	for _, reChild := range n.reChildren {
		output.WriteString(r.stringRec(reChild, l+1))
	}
	return output.String()
}

func (r *radix) String() string {
	return r.stringRec(r.root, 0)
}

// Add path(p) from node(n), if n is nil, insert directly on n, otherwise compare n's path with p, find common prefix,
// if prefix is part of n, split n's path with creating new n, and update current n's path with rest of prefix and add as new n's child.
// then create new node with the rest of p, add as new n's child.
// wildcard segment will be considered as single node.
func (r *radix) putRec(n *node, path string, handler func(ctx *Context)) (t *node) {
	//put whole path in new node if path is plain text, otherwise parse path to take the plain text part or single regex part
	if n == nil {
		if idx, eof := parsePref(path); eof {
			// the whole tail path in new node
			t = newNode(path)
			t.handler = handler
		} else {
			// put prefix in new node and the tail in the child level nodes.
			child := r.putRec(nil, path[idx+1:], handler)
			if child != nil {
				t = newNode(path[:idx+1])
				if child.rePatterns == nil {
					t.children[parseKey(child.part)] = child
				} else {
					t.reChildren = append(t.reChildren, child)
				}
			}
		}
	} else {
		ln, lp := len(n.part), len(path)
		min := ln
		if min > lp {
			min = lp
		}
		i, idx := 0, -1
		for i < min {
			if n.part[i] != path[i] {
				break
			}
			if n.part[i] == '/' {
				idx = i
			}
			i++
		}
		if i < ln {
			//there is tail of node path indeed, create new node for both common prefix
			t = r.putRec(nil, n.part[:idx+1], nil)
			n.part = n.part[idx+1:]
			t.children[parseKey(n.part)] = n
			if i == lp {
				t.handler = handler
			}
		} else {
			t = n
		}
		if i < lp {
			var tail *node
			//there is tail of path indeed, create new node for tail of new path
			tailPath := path[idx+1:]
			tailKey := parseKey(tailPath)
			if child, ok := n.children[tailKey]; ok {
				tail = r.putRec(child, tailPath, handler)
			} else if child, ok := n.getReChild(tailKey); ok {
				r.putRec(child, tailPath, handler)
			} else {
				tail = r.putRec(nil, tailPath, handler)
			}
			if tail != nil {
				if tail.rePatterns == nil {
					t.children[tailKey] = tail
				} else {
					t.reChildren = append(t.reChildren, tail)
				}
			}
		}
	}
	return
}

func (r *radix) put(path string, handler func(*Context)) (b bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if t := r.putRec(r.root, path, handler); t != nil {
		r.root = t
		r.size++
		b = true
	}
	return
}

func (r *radix) getRec(n *node, path string, params map[string]string) (t *node) {
	if n.rePatterns == nil || n.rePatterns.len() == 0 {
		// current node has plain text part, compare with path and call recursively with child and tail, then return found node.
		// find child firstly from plain text children map, if child is not found, loop call with every regex child with tail, find the first not null node.
		if tail, ok := strings.CutPrefix(path, n.part); ok {
			if len(tail) > 0 {
				if child, ok := n.children[parseKey(tail)]; ok {
					t = r.getRec(child, tail, params)
				} else {
					for _, reChild := range n.reChildren {
						if t = r.getRec(reChild, tail, params); t != nil {
							break
						}
					}
				}
			} else if n.handler != nil {
				t = n
			}
		}
	} else {
		if params == nil {
			params = make(map[string]string)
		}
	}
	return
}

// Get handler from cache by path, if it's not exist, get recursively from tree.
func (r *radix) get(path string) (func(*Context), map[string]string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	n, params := r.cache.get(path)
	if n == nil {
		if r.root == nil {
			return nil, nil
		}
		if n = r.getRec(r.root, path, params); n != nil {
			r.cache.put(path, n, params)
		} else {
			return nil, nil
		}
	}
	return n.handler, params
}

// Delete leaf node, then recursively delete parent node if it's alone
func (r *radix) deleteRec(n *node, p string) (b bool) {
	return true
}

// Delete root node if children lists are empty.
func (r *radix) delete(path string) (b bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.deleteRec(r.root, path) {
		r.size--
		if strings.Contains(path, "{") {
			r.cache.clear()
		} else {
			r.cache.delete(path)
		}
	}
	return false
}

func newRouter() *router {
	return &router{
		trees: make(map[string]*radix),
	}
}

func (r *router) clear() {
	for _, r := range r.trees {
		r.clear()
	}
}

func (r *router) len() (c int) {
	for _, r := range r.trees {
		c += r.len()
	}
	return
}

func (r *router) put(method string, path string, handler func(*Context)) {

	log.Printf("Put route %4s - %s", method, path)
	if strings.HasPrefix(path, "/") {
		panic("Path must begin with '/'!")
	}
	if handler == nil {
		panic("Handler function should not be nil!")
	}
	if _, ok := r.trees[method]; !ok {
		r.trees[method] = newRadix()
	}
	r.trees[method].put(path, handler)
}

func (r *router) get(method string, path string) (handler func(*Context)) {
	log.Printf("Get route %4s - %s", method, path)
	if path[0] != '/' {
		panic("Path must begin with '/'!")
	}
	if _, ok := r.trees[method]; ok {
		//return tree.get(path)
	}
	return nil
}

func (r *router) delete(method string, path string) (b bool) {
	log.Printf("Delete route %4s - %s", method, path)

	if _, ok := r.trees[method]; ok {
		//return tree.delete(path)
	}
	return false
}
