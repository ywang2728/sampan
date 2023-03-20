package sampan

import (
	"errors"
	"log"
	"regexp"
	"strings"
	"sync"
)

type (
	node struct {
		path     string
		part     string
		exps     map[string]*regexp.Regexp
		children map[string]*node
		params   map[string]string
		handler  func(*Context)
	}

	radix struct {
		root  *node
		count int
		mutex sync.RWMutex
	}

	router struct {
		trees    map[string]*radix
		handlers map[string]func(*Context)
	}
)

func formatPath(p string) {
	if p[0] != '/' {
		log.Fatalf("URL path formating error, #%v", errors.New(`invalid URL path:`+p))
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
				if strings.Contains(part[i:len(part)], "{") {
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
	n = new(node)
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
	n.children = make(map[string]*node)
	return
}

func newRadix() (r *radix) {
	r = new(radix)
	r.mutex = sync.RWMutex{}
	return
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
		r.count = 0
	}
}

func (r *radix) isEmpty() bool {
	if r == nil {
		return false
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	return r.count == 0
}

func (r *radix) size() int {
	if r == nil {
		return 0
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.root == nil {
		return 0
	}
	return r.count
}

/*func (r *radix) get(path string) (handler func(*Context)) {
	if r == nil {
		return nil
	}
	r.mutex.RLock()
	defer r.mutex.RUnlock()
	if r.root == nil {
		return nil
	}
	handler = r.root.handler
	if path != "/" {
		t := r.root.get(parse(path))
		if t == nil {
			return nil
		}
		handler = t.handler
	}
	return
}

func (r *radix) put(path string, handler func(*Context)) (b bool) {
	if r == nil {
		return false
	}
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.root == nil {
		r.root = newNode([]string{"/"})
	}
	t := r.root
	if path != "/" {
		r.root, t = r.root.put(parse(path))
	}
	if t != nil && len(t.path) == 0 {
		t.path = path
		t.handler = handler
		r.count++
	}
	return
}

func (r *radix) update(path string, handler func(*Context)) (b bool) {
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

func (r *radix) delete(path string) (b bool) {
	if r == nil {
		return false
	}
	parts := parse(path)
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if len(parts) == 1 {
		if len(r.root.children) != 0 {
			return false
		}
		r.root.handler = nil
		r.count--
		return true
	} else if r.root.delete(parts[1:]) {
		r.count--
		return true
	}
	return false
}*/

func newRouter() *router {
	return &router{
		trees:    make(map[string]*radix),
		handlers: make(map[string]func(*Context)),
	}
}

/*func (r *router) put(method string, path string, handler func(*Context)) {

	log.Printf("Put route %4s - %s", method, path)
	if path[0] != '/' {
		panic("Path must begin with '/'!")
	}
	if handler == nil {
		panic("Handler function should not be nil!")
	}
	if _, ok := r.trees[method]; !ok {
		r.trees[method] = newRadix()
	}
	r.trees[method].put(path, handler)

	key := fmt.Sprintf("%s-%s", method, path)
	r.handlers[key] = handler
}

func (r *router) get(method string, path string) (handler func(*Context)) {
	log.Printf("Get route %4s - %s", method, path)
	if path[0] != '/' {
		panic("Path must begin with '/'!")
	}
	if tree, ok := r.trees[method]; ok {
		return tree.get(path)
	}
	return nil
}

func (r *router) delete(method string, path string) (b bool) {
	log.Printf("Delete route %4s - %s", method, path)

	if tree, ok := r.trees[method]; ok {
		return tree.delete(path)
	}
	return false
}*/
