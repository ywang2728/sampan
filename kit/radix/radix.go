package radix

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

type (
	Key[K comparable] interface {
		fmt.Stringer
		Value() K
		// Match the current key with input key, return common prefix, and tails of current key and input key.
		Match(K) (KeyIterable[K], KeyIterable[K], KeyIterable[K], *map[K]K)
	}

	KeyIterable[K comparable] interface {
		hasNext() bool
		Next() Key[K]
	}

	node[K comparable, V any] struct {
		k     Key[K]
		v     *V
		nodes []*node[K, V]
	}

	Radix[K comparable, V any] struct {
		size int
		root *node[K, V]
		// Func to build Key Iterator, the Key struct could be String, Wildcard, or Regex.
		buildKeyIter func(K) KeyIterable[K]
		mtx          sync.RWMutex
	}
)

func New[K comparable, V any](keyIterFunc func(K) KeyIterable[K]) *Radix[K, V] {
	return &Radix[K, V]{
		buildKeyIter: keyIterFunc,
	}
}

func (r *Radix[K, V]) newNode(k Key[K]) (n *node[K, V]) {
	n = &node[K, V]{
		k:     k,
		nodes: []*node[K, V]{},
	}
	return
}

func (r *Radix[K, V]) Clear() {
	if r != nil {
		r.mtx.Lock()
		defer r.mtx.Unlock()
		r.root = nil
		r.size = 0
	}
}

func (r *Radix[K, V]) Len() int {
	if r == nil {
		return 0
	}
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	if r.root == nil {
		return 0
	}
	return r.size
}
func (r *Radix[K, V]) stringRec(n *node[K, V], l int) string {
	output := strings.Builder{}
	output.WriteString(strings.Repeat("#", l))
	output.WriteString(fmt.Sprintf(" %p : %+v\n", n, n))
	for _, child := range n.nodes {
		output.WriteString(r.stringRec(child, l+1))
	}
	return output.String()
}

func (r *Radix[K, V]) String() string {
	return r.stringRec(r.root, 0)
}

func (r *Radix[K, V]) put(k K, v V) (b bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	if nr := r.putRec(r.root, r.buildKeyIter(k), &v); nr != nil {
		r.root = nr
		r.size++
		b = true
	}
	return
}

func (r *Radix[K, V]) putRec(n *node[K, V], keyIter KeyIterable[K], value *V) (nn *node[K, V]) {
	if keyIter.hasNext() {
		key := keyIter.Next()
		if n == nil {
			nn = r.newNode(key)
			if keyIter.hasNext() {
				nn = r.putRec(nn, keyIter, value)
			} else {
				nn.v = value
			}
		} else {
			c, tn, tp, _ := n.k.Match(key.Value())
			if c != nil && c.hasNext() {
				if tn != nil && tn.hasNext() {
					nn = r.putRec(nil, c, nil)
					n.k = tn.Next()
					nn.nodes = append(nn.nodes, n)
				} else {
					nn = n
				}
			}
			if tp != nil && tp.hasNext() {
				var tpn *node[K, V]
				for _, child := range n.nodes {
					if cc, _, _, _ := child.k.Match(tp.Next().Value()); cc != nil && cc.hasNext() {
						tpn = r.putRec(child, tp, value)
						break
					}
				}
				if tpn == nil {
					tpn = r.putRec(nil, tp, value)
				}
				nn.nodes = append(nn.nodes, tpn)
			} else if nn.v == nil {
				nn.v = value
			} else {
				log.Fatalf("Input key error, %#v\t", errors.New(fmt.Sprintf(`duplicated key:%#v`, key)))
			}
		}
	}
	return
}
