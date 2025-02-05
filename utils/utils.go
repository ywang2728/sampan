package utils

import "github.com/ywang2728/sampan/ds/stack"

func indexNth(key string, char uint8, n int) int {
	for occur, i := 0, 0; i < len(key); i++ {
		if key[i] == char {
			if occur++; occur == n {
				return i
			}
		}
	}
	return -1
}

func infix2Suffix(infix *string) string {
	s := stack.NewMutexStack[string]()

}
