package sampan

import (
	"log"
	"net/http"
)

type Server struct {
	rg *RouterGroup
}

func New() (s *Server) {
	s = &Server{}
	s.rg = NewRouterGroup("", newRouter())
	return
}

func (s *Server) Group(prefix string) (g *RouterGroup) {
	return s.rg.Group(prefix)
}

func (s *Server) DeleteGroup(prefix string) {
	s.rg.DeleteGroup(prefix)
}

func (s *Server) PreMiddlewares(middlewares ...func(*Context)) {
	s.rg.PreMiddlewares(middlewares...)
}

func (s *Server) PostMiddlewares(middlewares ...func(*Context)) {
	s.rg.PostMiddlewares(middlewares...)
}

func (s *Server) GetRoute(method string, path string) (handlerChain []func(*Context), params map[string]string) {
	return s.rg.GetRoute(method, path)
}

func (s *Server) PutRoute(method string, path string, handler func(*Context)) {
	s.rg.PutRoute(method, path, handler)
}

func (s *Server) GET(path string, handler func(*Context)) {
	s.PutRoute(http.MethodGet, path, handler)
}

func (s *Server) POST(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPost, path, handler)
}

func (s *Server) PUT(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPut, path, handler)
}

func (s *Server) PATCH(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPatch, path, handler)
}

func (s *Server) DELETE(path string, handler func(*Context)) {
	s.PutRoute(http.MethodDelete, path, handler)
}

func (s *Server) HEAD(path string, handler func(*Context)) {
	s.PutRoute(http.MethodHead, path, handler)
}

func (s *Server) OPTIONS(path string, handler func(*Context)) {
	s.PutRoute(http.MethodOptions, path, handler)
}

func (s *Server) handle(c *Context) {
	log.Printf("Handle %4s - %s", c.Method, c.Path)
	handlerChain, params := s.rg.GetRoute(c.Method, c.Path)
	if len(handlerChain) > 0 {
		c.setParams(params)
		for _, handler := range handlerChain {
			handler(c)
		}
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.handle(newContext(w, req))
}

func (s *Server) Listen(addr string) (err error) {
	return http.ListenAndServe(addr, s)
}
