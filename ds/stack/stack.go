package stack

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

type (
	Stacker[E any] interface {
		Push(e E)
		Pop() E
	}

	element[E any] struct {
		value E
		next  unsafe.Pointer
	}

	MutexStack[E any] struct {
		top unsafe.Pointer
		len uint64
		mtx sync.RWMutex
	}

	CasStack[E any] struct {
		top unsafe.Pointer
		len uint64
	}
)

//Mutex Stack

func NewMutexStack[E any]() *MutexStack[E] {
	return &MutexStack[E]{}
}

func (ms *MutexStack[E]) Push(e E) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	elem := unsafe.Pointer(&element[E]{value: e})
	if ms.top != nil {
		(*element[E])(elem).next = ms.top
	}
	ms.top = elem
	ms.len++
}

func (ms *MutexStack[E]) Pop() (value E, ok bool) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	if ms.top != nil {
		value = (*element[E])(ms.top).value
		ms.top = (*element[E])(ms.top).next
		ms.len--
		ok = true
	}
	return
}

func (ms *MutexStack[E]) Peek() (value E, ok bool) {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()
	if ms.top != nil {
		value = (*element[E])(ms.top).value
		ok = true
	}
	return
}

func (ms *MutexStack[E]) IsEmpty() bool {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()
	return ms.top == nil
}

func (ms *MutexStack[E]) String() string {
	var sb strings.Builder
	sb.WriteString("MutexStack[")
	for curr := ms.top; curr != nil; curr = (*element[E])(curr).next {
		sb.WriteString(fmt.Sprintf("%v ", (*element[E])(curr).value))
	}
	return strings.TrimRight(sb.String(), " ") + "]"
}

// CAS Stack

func NewCasStack[E any]() *CasStack[E] {
	return &CasStack[E]{}
}

func (cs *CasStack[E]) Push(e E) {
	elem := &element[E]{value: e}
	for {
		if cs.top == nil {
			atomic.StorePointer(&cs.top, unsafe.Pointer(elem))
			return
		}
		old := atomic.LoadPointer(&cs.top)
		elem.next = old
		if atomic.CompareAndSwapPointer(&cs.top, old, unsafe.Pointer(elem)) {
			atomic.AddUint64(&cs.len, 1)
			return
		}
	}
}

func (cs *CasStack[E]) Pop() (value E, ok bool) {
	for {
		old := atomic.LoadPointer(&cs.top)
		if old != nil {
			oldElem := (*element[E])(old)
			next := atomic.LoadPointer(&oldElem.next)
			if atomic.CompareAndSwapPointer(&cs.top, old, next) {
				atomic.AddUint64(&cs.len, ^uint64(0))
				return oldElem.value, true
			}
		}
	}
}

func (cs *CasStack[E]) Peek() (value E, ok bool) {
	for {
		if cs.top == nil {
			return value, false
		} else {
			old := atomic.LoadPointer(&cs.top)
			return (*element[E])(old).value, true
		}
	}
}

func (cs *CasStack[E]) IsEmpty() bool {
	for {
		return atomic.LoadPointer(&cs.top) == nil
	}
}

func (cs *CasStack[E]) String() string {
	var sb strings.Builder
	sb.WriteString("MutexStack[")
	for curr := atomic.LoadPointer(&cs.top); curr != nil; curr = (*element[E])(curr).next {
		sb.WriteString(fmt.Sprintf("%v ", (*element[E])(curr).value))
	}
	return strings.TrimRight(sb.String(), " ") + "]"
}
