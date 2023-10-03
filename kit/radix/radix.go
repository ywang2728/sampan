package radix

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
)

type (
	Parser[K comparable] interface {
		Compare(K, K) (K, K, K)
	}

	node[K comparable, V any] struct {
		key    K
		value  V
		nodes  []*node[K, V]
		parser Parser[K]
	}

	Radix[K comparable, V any] struct {
		root      *node[K, V]
		size      int
		newParser func(K) (Parser[K], K, K)
		mtx       sync.RWMutex
	}
)

func New[K comparable, V any](np func(K) (Parser[K], K, K)) *Radix[K, V] {
	return &Radix[K, V]{
		newParser: np,
	}
}

func (r *Radix[K, V]) newNode(k K, p Parser[K]) (n *node[K, V]) {
	n = &node[K, V]{
		key:    k,
		nodes:  []*node[K, V]{},
		parser: p,
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
	if nr := r.putRec(r.root, k, v); nr != nil {
		r.root = nr
		r.size++
		b = true
	}
	return
}

func (r *Radix[K, V]) putRec(n *node[K, V], k K, v V) (nn *node[K, V]) {
	if n == nil {
		parser, key, tail := r.newParser(k)
		nn = r.newNode(key, parser)
		if tail == nil {
			nn.value = v
		} else {
			nn = r.putRec(nn, tail, v)
		}
	} else {
		c, tn, tp := n.parser.Compare(n.key, k)
		if c != nil {
			if tn != nil {
				nn = r.putRec(nil, c, nil)
				n.key = tn
				nn.nodes = append(nn.nodes, n)
			} else {
				nn = n
			}
		}
		if tp != nil {
			var np *node[K, V]
			for _, child := range n.nodes {
				if cc, _, _ := child.parser.Compare(child.key, tp); cc != nil {
					np = r.putRec(child, tp, v)
					break
				}
			}
			if np == nil {
				np = r.putRec(nil, tp, v)
			}
			nn.nodes = append(nn.nodes, np)
		} else {
			if nn.value == nil {
				nn.value = v
			} else {
				log.Fatalf("Input key error, %#v\t", errors.New(fmt.Sprintf(`duplicated key:%#v`, k)))
			}
		}
	}
	return
}
