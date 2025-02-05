package stack

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

type (
	Stacker[E any] interface {
		Push(e E)
		Pop() E
	}

	element[E any] struct {
		prev  unsafe.Pointer
		value E
		next  unsafe.Pointer
	}

	MutexStack[E any] struct {
		head *element[E]
		tail *element[E]
		mtx  sync.RWMutex
	}

	CasStack[E any] struct {
		head *element[E]
	}
)

func NewMutexStack[E any]() *MutexStack[E] {
	return &MutexStack[E]{}
}

func NewCasStack[E any]() *CasStack[E] {
	return &CasStack[E]{}
}

func (ms *MutexStack[E]) Push(e E) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	elem := &element[E]{value: e}
	if ms.tail == nil {
		ms.head = elem
	} else {
		elem.prev = unsafe.Pointer(ms.tail)
		ms.tail.next = unsafe.Pointer(elem)
	}
	ms.tail = elem
}

func (ms *MutexStack[E]) Pop() (value E, ok bool) {
	ms.mtx.Lock()
	defer ms.mtx.Unlock()
	if ms.tail != nil {
		ok = true
		value = ms.tail.value
		ms.tail = (*element[E])(ms.tail.prev)
		if ms.tail == nil {
			ms.head = nil
		} else {
			ms.tail.next = nil
		}
	}
	return
}

func (ms *MutexStack[E]) Peek() (value E, ok bool) {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()
	if ms.tail != nil {
		ok = true
		value = ms.tail.value
	}
	return
}

func (ms *MutexStack[E]) IsEmpty() bool {
	ms.mtx.RLock()
	defer ms.mtx.RUnlock()
	return ms.head == nil
}

func (ms *MutexStack[e]) String() string {
	var sb strings.Builder
	sb.WriteString("MutexStack[")
	for curr := ms.head; curr != nil; curr = (*element[e])(curr.next) {
		sb.WriteString(fmt.Sprintf("%v ", curr.value))
	}
	return strings.TrimRight(sb.String(), " ") + "]"
}
