package main

import (
	"github.com/ywang2728/sampan/web"
	"net/http"
)

func main() {
	s := web.New()
	s.GET("/hello", func(ctx *web.Context) {
		ctx.String(http.StatusOK, "hello world!")
	})
	s.GET("/index", func(ctx *web.Context) {
		ctx.HTML(http.StatusOK, "<h1>Index Page</h1>")
	})
	g := s.Group("/v1")
	g.PreMiddlewares(func(ctx *web.Context) {
		ctx.Writer.Write([]byte("<p>hahaha</p><br>"))
	})
	g.GET("/haha", func(ctx *web.Context) {
		ctx.HTML(http.StatusOK, "<h1>HAHAHA</h1>")
	})
	g.PutStaticRoute("/tmp", "/tmp")
	s.Listen(":12345")
}
