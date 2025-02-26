package utils

import (
	"github.com/ywang2728/sampan/ds/stack"
	"slices"
	"strings"
	"unicode"
)

func IndexNth(key string, char uint8, n int) int {
	for occur, i := 0, 0; i < len(key); i++ {
		if key[i] == char {
			if occur++; occur == n {
				return i
			}
		}
	}
	return -1
}

func Infix2Suffix(infix *string) (suffix []string) {
	priority := map[rune]int8{'+': 2, '-': 2, '*': 3, '/': 3, '(': 1}
	opt := []rune{'+', '-', '*', '/', '(', ')'}
	optWithoutBrace := []rune{'+', '-', '*', '/', '('}
	sign := []rune{'+', '-'}
	point := '.'

	s := stack.NewCasStack[rune]()
	inf := []rune(*infix)
	l := len(inf)
	var num strings.Builder

	for i := 0; i < l; i++ {
		if unicode.IsDigit(inf[i]) ||
			(slices.Contains(sign, inf[i]) && (i == 0 || slices.Contains(optWithoutBrace, inf[i-1]))) ||
			(point == inf[i] && (i != l-1 && unicode.IsDigit(inf[i-1]) && unicode.IsDigit(inf[i+1]))) {
			if '+' != inf[i] {
				num.WriteRune(inf[i])
			}
		} else if slices.Contains(opt, inf[i]) {
			if num.Len() != 0 {
				suffix = append(suffix, num.String())
				num.Reset()
			}
			switch inf[i] {
			case '(':
				s.Push(inf[i])
			case '*', '/', '+', '-':
				if s.IsEmpty() {
					s.Push(inf[i])
				} else {
					for v, ok := s.Peek(); ok && priority[v] >= priority[inf[i]]; v, ok = s.Peek() {
						suffix = append(suffix, string(v))
						s.Pop()
					}
					s.Push(inf[i])
				}
			default:
				for v, ok := s.Pop(); ok && v != '('; v, ok = s.Pop() {
					suffix = append(suffix, string(v))
				}
			}
		} else {
			panic("Invalid char!")
		}
	}
	if num.Len() != 0 {
		suffix = append(suffix, num.String())
	}
	for v, ok := s.Pop(); ok; v, ok = s.Pop() {
		suffix = append(suffix, string(v))
	}

	return
}
