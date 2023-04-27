package main

import (
	"github.com/ywang2728/sampan"
	"net/http"
)

func main() {
	s := sampan.New()
	s.GET("/hello", func(ctx *sampan.Context) {
		ctx.String(http.StatusOK, "hello world!")
	})
	s.GET("/index", func(ctx *sampan.Context) {
		ctx.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	g := s.Group("/v1")
	g.PreMiddlewares(func(ctx *sampan.Context) {
		ctx.Writer.Write([]byte("<p>hahaha</p><br>"))
	})
	g.GET("/haha", func(ctx *sampan.Context) {
		ctx.HTML(http.StatusOK, "<h1>HAHAHA</h1>")
	})
	s.Listen(":12345")
}
