package sampan

import (
	"container/list"
	"errors"
	"log"
	"regexp"
	"strings"
	"sync"
)

const (
	LruCapacity int = 255
)

type (
	node struct {
		path     string
		part     string
		exps     map[string]*regexp.Regexp
		params   map[string]string
		handler  func(*Context)
		children map[string]*node
	}

	lruNode struct {
		path string
		node *node
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

func splitPath(path string) (parts []string) {
	part := strings.Builder{}
	for i, t := 0, 0; i < len(path); i++ {
		part.WriteString(string(path[i]))
		if path[i] == '{' {
			t++
		} else if path[i] == '}' {
			t--
		} else if path[i] == '/' && t == 0 {
			parts = append(parts, part.String())
			part.Reset()
		}
	}
	if part.Len() > 0 {
		parts = append(parts, part.String())
	}
	return
}

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

func (l *lru) put(path string, node *node) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if e, ok := l.paths[path]; ok {
		l.nodes.MoveToFront(e)
	} else {
		if l.nodes.Len() == l.cap {
			delete(l.paths, l.nodes.Remove(l.nodes.Back()).(*lruNode).path)
		}
		l.paths[path] = l.nodes.PushFront(&lruNode{path: path, node: node})
	}
}

func (l *lru) get(path string) (n *node) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if e, ok := l.paths[path]; ok {
		l.nodes.MoveToFront(e)
		n = e.Value.(*lruNode).node
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

func parseExps(part string) (matches []string) {
	expKeyRe := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*$`)
	var sb strings.Builder
	t := 0
	for i := strings.Index(part, "{"); i < len(part); i++ {
		if part[i] == '{' {
			t++
			if t != 1 {
				sb.WriteString("{")
			}
		} else if part[i] == '}' {
			t--
			if t != 0 {
				sb.WriteString("}")
			}
			if t == 0 {
				s := sb.String()
				if strings.Contains(s, ":") {
					if strings.HasPrefix(s, ":") || strings.HasSuffix(s, ":") ||
						!expKeyRe.MatchString(s[0:strings.Index(s, ":")]) {
						log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+s))
					}
				} else {
					if !expKeyRe.MatchString(s) {
						log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+s))
					}
				}
				matches = append(matches, s)
				if strings.Contains(part[i:], "{") {
					sb.Reset()
				} else {
					break
				}
			}
		} else if t != 0 {
			sb.WriteString(string(part[i]))
		}
	}
	if t != 0 {
		log.Fatalf("Expression parsing error, #%v", errors.New(`invalid expression:`+part))
	}
	return
}

func newNode(part string) (n *node) {
	n = &node{
		children: make(map[string]*node),
	}
	if len(part) > 0 {
		n.part = part
		if strings.Contains(part, "{") {
			n.exps = make(map[string]*regexp.Regexp)
			for _, exp := range parseExps(part) {
				k, v, ok := strings.Cut(exp, ":")
				if ok {
					n.exps[k] = regexp.MustCompile(v)
				} else {
					n.exps[k] = regexp.MustCompile(`\S*`)
				}
			}
		}
	}
	return
}

func newRadix() *radix {
	return &radix{
		cache: newLru(LruCapacity),
	}
}

/*func (n *node) put(parts []string) (r *node, t *node) {
	nl, pl := len(n.parts), len(parts)
	min := nl
	if pl < nl {
		min = pl
	}
	dp := -1
	for i := 0; i < min; i++ {
		if n.parts[i] != parts[i] {
			dp = i
			break
		}
	}
	if dp < 0 {
		if nl == pl {
			return n, n
		} else if nl < pl {
			if child, ok := n.children[parts[nl]]; ok {
				return child.put(parts[nl:])
			} else {
				t = newNode(parts[nl:])
				n.children[parts[nl]] = t
				return n, t
			}
		} else {
			r = newNode(parts)
			n.parts = n.parts[pl:]
			r.children[n.parts[pl]] = n
			return r, r
		}
	} else {
		r = newNode(n.parts[0:dp])
		n.parts = n.parts[dp:]
		r.children[n.parts[dp]] = n
		t = newNode(parts[dp:])
		r.children[parts[dp]] = t
		return r, t
	}
}*/

/*
	func (n *node) get(path string) (t *node) {
		nl, pl := len(n.part), len(path)
		if nl > pl {
			return nil
		}
		p := lcp(&n.part, &path)
		if nl == pl {
			if p != nl {
				return nil
			}
			for i := 0; i < nl; i++ {
				if n.parts[i] != parts[i] {
					return nil
				}
			}
			return n
		}
		if child, ok := n.children[parts[nl]]; ok {
			for i := 0; i < nl; i++ {
				if n.parts[i] != parts[i] {
					return nil
				}
			}
			return child.get(parts[nl:])
		}
		return nil
	}

	func (n *node) delete(parts []string) (b bool) {
		if child, ok := n.children[parts[0]]; ok {
			cl, pl := len(child.parts), len(parts)
			if cl <= pl {
				dp := -1
				for i := 0; i < cl; i++ {
					if child.parts[i] != parts[i] {
						dp = i
					}
				}
				if dp < 0 {
					if len(child.children) == 0 {
						delete(n.children, parts[0])
						return true
					}
					return false
				} else {
					return child.delete(parts[dp:])
				}
			} else {
				return false
			}
		}
		return false
	}
*/
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

// Add path(p) from node(n), if n is nil, insert directly on n, otherwise compare n's path with p, find common prefix,
// if prefix is part of n, split n's path with creating new n, and update current n's path with rest of prefix and add as new n's child.
// then create new node with the rest of p, add as new n's child.
// wildcard segment will be considered as single node.
func (r *radix) putRec(n *node, p string) (nr, t *node) {
	return nil, nil
}

func (r *radix) put(path string, handler func(*Context)) (b bool) {
	if r == nil {
		return false
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	var t *node
	r.root, t = r.putRec(r.root, path)
	if t != nil && len(t.path) == 0 {
		t.path = path
		t.handler = handler
		r.size++
	}
	return
}

func (r *radix) getRec(n *node, p string) (t *node) {
	return nil
}

func (r *radix) get(path string) func(*Context) {
	if r == nil {
		return nil
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	n := r.cache.get(path)
	if n == nil {
		if r.root == nil {
			return nil
		}
		if n = r.getRec(r.root, path); n != nil {
			r.cache.put(path, n)
		} else {
			return nil
		}
	}
	//TODO: handle n.params
	return n.handler
}

func (r *radix) deleteRec(n *node, p string) (b bool) {
	return true
}

func (r *radix) delete(path string) (b bool) {
	if r == nil {
		return false
	}
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

/*func (r *radix) update(path string, handler func(*Context)) (b bool) {
	if r == nil {
		return false
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.root == nil {
		return false
	}
	t := r.root
	if path != "/" {
		t = r.root.get(parse(path))
	}
	if t.path == path && t.handler != nil {
		t.handler = handler
		return true
	}
	return false
}
*/

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
