package stack

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

type (
	Stack[E any] struct {
		l *list.List
		sync.RWMutex
	}
)

func New[E any]() *Stack[E] {
	return &Stack[E]{
		l: list.New(),
	}
}

func (s *Stack[E]) Push(e E) {
	s.Lock()
	defer s.Unlock()
	s.l.PushFront(e)
}

func (s *Stack[E]) Pop() (e E, ok bool) {
	s.Lock()
	defer s.Unlock()
	elem := s.l.Front()
	if elem != nil {
		e = s.l.Remove(elem).(E)
		ok = true
	}
	return
}

func (s *Stack[E]) IsEmpty() bool {
	s.RLock()
	defer s.RUnlock()
	return s.l.Front() == nil
}

func (s *Stack[E]) Peek() (e E) {
	s.RLock()
	defer s.RUnlock()
	elem := s.l.Front()
	if elem != nil {
		e = elem.Value.(E)
	}
	return
}

func (s *Stack[e]) String() string {
	var sb strings.Builder
	sb.WriteString("Stack[")
	for e := s.l.Front(); e != nil; e = e.Next() {
		sb.WriteString(fmt.Sprintf("%+v ", e.Value))
	}
	return strings.TrimRight(sb.String(), " ") + "]"
}
