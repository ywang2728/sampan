package web2

import (
	"errors"
	"fmt"
)

func (r *Radix[K, V]) put(k K, v V) (b bool) {
	r.Lock()
	defer r.Unlock()
	if nr := r.putRec(r.root, r.buildKeyIterator(k), &v); nr != nil {
		r.root = nr
		r.size++
		b = true
	}
	return
}

func (r *Radix[K, V]) putRec(n *node[K, V], keyIter KeyIterator[K], value *V) (nn *node[K, V]) {
	if keyIter.HasNext() {
		key := keyIter.Next()
		if n == nil {
			nn = r.newNode(key)
			if keyIter.HasNext() {
				nn = r.putRec(nn, keyIter, value)
			} else {
				nn.v = value
			}
		} else {
			c, tn, tp, _ := n.k.Match(key)
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
				tpk := tp.Peek()
				for i := 0; i < len(n.nodes); i++ {
					if cc, _, _, _ := n.nodes[i].k.Match(tpk); cc != nil && cc.HasNext() {
						tpn = r.putRec(n.nodes[i], tp, value)
						n.nodes[i] = tpn
						break
					}
				}
				if tpn == nil {
					tpn = r.putRec(nil, tp, value)
					nn.nodes = append(nn.nodes, tpn)
				}
			} else if nn.v == nil {
				nn.v = value
			} else {
				log.Fatalf("Input key error, %#v\t", errors.New(fmt.Sprintf(`duplicated key:%#v`, key)))
			}
		}
	}
	return
}
