package sampan

import (
	"container/list"
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	LruCapacity = 255
	ReDelimBgn  = '{'
	ReDelimEnd  = '}'
)

type (
	reDelim struct {
		cnt atomic.Int32
	}

	rePattern struct {
		raw      string
		compiled *regexp.Regexp
	}

	node struct {
		part       string
		rePatterns []*rePattern
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

	RouterGroup struct {
		prefix      string
		middlewares []func(*Context)
		router      *router
		children    map[string]*RouterGroup
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

func parseRePatterns(part string) (patterns []*rePattern) {
	patterns = []*rePattern{}
	for len(part) != 0 {
		if i := strings.Index(part, string(ReDelimBgn)); i == -1 {
			patterns, part = append(patterns, &rePattern{part, nil}), ""
		} else if i == 0 {
			delim := newReDelim()
			for ; i < len(part); i++ {
				if part[i] == ReDelimBgn {
					delim.open()
				} else if part[i] == ReDelimEnd && delim.close() {
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
					patterns, part = append(patterns, &rePattern{before, regexp.MustCompile(before[1 : l-1])}), after
					break
				}
			}
			if !delim.closed() {
				log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
			}
		} else {
			patterns = append(patterns, &rePattern{part[:i], nil})
			part = part[i:]
		}
	}
	return
}

// parse the prefix from path for one node base on the difference of regex segment and plain text segment.
func parsePrefix(path string) (idx int, eof bool) {
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

func parseCommonPrefix(part, path string) (idx, i, ln, lp int) {
	ln, lp = len(part), len(path)
	min := ln
	if min > lp {
		min = lp
	}
	for idx = -1; i < min; i++ {
		if part[i] != path[i] {
			break
		}
		if part[i] == '/' {
			idx = i
		}
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

func newNode(part string) (n *node) {
	n = &node{
		children:   make(map[string]*node),
		reChildren: make([]*node, 0),
	}
	if len(part) > 0 {
		n.part = part
		if strings.Contains(part, string(ReDelimBgn)) {
			n.rePatterns = parseRePatterns(part)
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
		if idx, eof := parsePrefix(path); eof {
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
		idx, i, ln, lp := parseCommonPrefix(n.part, path)
		if i < ln || (i == ln && idx != i-1) {
			//there is tail of node path indeed, create new node for both common prefix
			t = r.putRec(nil, n.part[:idx+1], nil)
			n.part = n.part[idx+1:]
			t.children[parseKey(n.part)] = n
		} else {
			t = n
		}
		if idx == lp-1 {
			if t.handler == nil {
				t.handler = handler
			} else {
				log.Fatalf("Input path error, #%v", errors.New(`duplicated path:`+path))
			}
		}
		if i < lp || (i == lp && idx != i-1) {
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
	isMatched := true
	before := strings.Builder{}
	if n.rePatterns == nil {
		before.WriteString(n.part)
	} else {
		for i, p, l := 0, path, len(n.rePatterns); i < l && isMatched; i++ {
			ptn := n.rePatterns[i]
			if ptn.compiled == nil {
				if p, isMatched = strings.CutPrefix(p, ptn.raw); isMatched {
					before.WriteString(ptn.raw)
				}
			} else {
				if loc := ptn.compiled.FindStringIndex(p); loc != nil && loc[0] == 0 {
					if i != l-1 || loc[1] == len(p) || p[loc[1]] == '/' {
						toBeMatched := p[loc[0]:loc[1]]
						if p, isMatched = strings.CutPrefix(p, toBeMatched); isMatched {
							before.WriteString(toBeMatched)
							if names := ptn.compiled.SubexpNames(); len(names) > 1 {
								for i, match := range ptn.compiled.FindStringSubmatch(toBeMatched) {
									if len(names[i]) != 0 {
										params[names[i]] = match
									}
								}
							}
						}
					} else {
						isMatched = false
					}
				} else {
					isMatched = false
				}
			}
		}
	}
	if after, ok := strings.CutPrefix(path, before.String()); isMatched && ok {
		if len(after) > 0 {
			if child, ok := n.children[parseKey(after)]; ok {
				t = r.getRec(child, after, params)
			} else {
				for _, reChild := range n.reChildren {
					if t = r.getRec(reChild, after, params); t != nil {
						break
					}
				}
			}
		} else if n.handler != nil {
			t = n
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
		params := make(map[string]string)
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
	fmt.Println(n, p)
	return true
}

// Delete root node if children lists are empty.
func (r *radix) delete(path string) (b bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.deleteRec(r.root, path) {
		if len(r.root.children) == 0 && len(r.root.reChildren) == 0 {
			r.root = nil
			r.size = 0
			r.cache.clear()
		} else {
			r.size--
			if strings.Contains(path, "{") {
				r.cache.clear()
			} else {
				r.cache.delete(path)
			}
		}
		b = true
	}
	return
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

func (r *router) get(method string, path string) (handler func(*Context), params map[string]string) {
	log.Printf("Get route %4s - %s", method, path)
	if path[0] != '/' {
		panic("Path must begin with '/'!")
	}
	if tree, ok := r.trees[method]; ok {
		handler, params = tree.get(path)
	}
	return
}

func (r *router) delete(method string, path string) (b bool) {
	log.Printf("Delete route %4s - %s", method, path)
	if path[0] != '/' {
		panic("Path must begin with '/'!")
	}
	if tree, ok := r.trees[method]; ok {
		b = tree.delete(path)
	}
	return
}

func newDefaultGroup(r *router, middlewares []func(*Context), children map[string]*RouterGroup) *RouterGroup {
	return &RouterGroup{prefix: "", router: r, middlewares: middlewares, children: children}
}
func (rg *RouterGroup) RouterGroup(prefix string, middlewares []func(*Context), children map[string]*RouterGroup) (g *RouterGroup) {
	if _, ok := rg.children[prefix]; ok {
		log.Fatalf("Create RouterGroup error, #%v", errors.New(`duplicated group prefix:`+prefix))
	} else {
		g = &RouterGroup{prefix: prefix, router: rg.router, middlewares: middlewares, children: children}
		rg.children[prefix] = g
	}
	return
}

func (rg *RouterGroup) GetRoute(method string, path string) (handler func(*Context), params map[string]string) {
	return rg.router.get(method, path)
}

func (rg *RouterGroup) AddRoute(method string, path string, handler func(*Context)) {
	rg.router.put(method, path, handler)
}

func (rg *RouterGroup) GET(path string, handler func(*Context)) {
	rg.AddRoute(http.MethodGet, path, handler)
}

func (rg *RouterGroup) POST(path string, handler func(*Context)) {
	rg.AddRoute(http.MethodPost, path, handler)
}

func (rg *RouterGroup) PUT(path string, handler func(*Context)) {
	rg.AddRoute(http.MethodPut, path, handler)
}

func (rg *RouterGroup) DELETE(path string, handler func(*Context)) {
	rg.AddRoute(http.MethodDelete, path, handler)
}
