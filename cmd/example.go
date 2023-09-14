package main

import (
	"fmt"
	hahaha "github.com/ywang2728/sampan/haha"
	"runtime"
)

func init() {
	fmt.Printf("Map: %v\n", m)
	info = fmt.Sprintf("OS: %s, Arch: %s", runtime.GOOS, runtime.GOARCH)
	defer func() { fmt.Println("from init") }()
}

var m = map[int]string{1: "a", 2: "b"}

var info string

func main() {
	hahaha2.Haha2()
	hahaha.Ppp()
	println("main")
	println(info)
	defer func() { fmt.Println("fin de main") }()
}
