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
		// MatchIterator for putting node by Key, return common KeyIterator, current key tail KeyIterator and new putting KeyIterator
		MatchIterator(KeyIterator[K]) (KeyIterator[K], KeyIterator[K], KeyIterator[K])
		// Match for getting node by Key, return bool for matched, K for tail and
		Match(K) (K, map[K]K, bool)
	}

	KeyIterator[K comparable] interface {
		Reset()
		HasNext() bool
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
		newKeyIterator func(K) KeyIterator[K]
		sync.RWMutex
	}
)

func (n *node[K, V]) String() string {
	return fmt.Sprintf("&{k:%+v, v:%+v, nodes: %+v", n.k, n.v, n.nodes)
}

func New[K comparable, V any](newKeyIterFunc func(K) KeyIterator[K]) *Radix[K, V] {
	return &Radix[K, V]{
		newKeyIterator: newKeyIterFunc,
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
		r.Lock()
		defer r.Unlock()
		r.root = nil
		r.size = 0
	}
}

func (r *Radix[K, V]) Len() int {
	if r == nil {
		return 0
	}
	r.RLock()
	defer r.RUnlock()
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
	r.Lock()
	defer r.Unlock()
	if nr := r.putRec(r.root, r.newKeyIterator(k), &v); nr != nil {
		r.root = nr
		r.size++
		b = true
	}
	return
}

func (r *Radix[K, V]) putRec(n *node[K, V], ki KeyIterator[K], v *V) (nn *node[K, V]) {
	if ki.HasNext() {
		if n == nil {
			nn = r.newNode(ki.Next())
			if ki.HasNext() {
				nn = r.putRec(nn, ki, v)
			} else {
				nn.v = v
			}
		} else {
			c, tn, tp := n.k.MatchIterator(ki)
			if c != nil && c.HasNext() {
				if tn != nil && tn.HasNext() {
					nn = r.putRec(nil, c, nil)
					n.k = tn.Next()
					nn.nodes = append(nn.nodes, n)
				} else {
					nn = n
				}
			}
			if tp != nil && tp.HasNext() {
				var tpn *node[K, V]
				for i := 0; i < len(n.nodes); i++ {
					tp.Reset()
					if cc, _, _ := n.nodes[i].k.MatchIterator(tp); cc != nil && cc.HasNext() {
						tpn = r.putRec(n.nodes[i], tp, v)
						n.nodes[i] = tpn
						break
					}
				}
				if tpn == nil {
					tpn = r.putRec(nil, tp, v)
					nn.nodes = append(nn.nodes, tpn)
				}
			} else if nn.v == nil {
				nn.v = v
			} else {
				log.Fatalf("Input key error, %#v\t", errors.New(fmt.Sprintf(`duplicated key:%#v`, ki.Next())))
			}
		}
	}
	return
}
