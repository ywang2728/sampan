package sampan

import (
	"net/http"
)

type Sampan struct {
	rg *RouterGroup
}

func New() (s *Sampan) {
	s = &Sampan{}
	s.rg = NewRouterGroup("", newRouter())
	return
}

func (s *Sampan) Group(prefix string) (g *RouterGroup) {
	return s.rg.Group(prefix)
}

func (s *Sampan) DeleteGroup(prefix string) {
	s.rg.DeleteGroup(prefix)
}

func (s *Sampan) PreMiddlewares(middlewares ...func(*Context)) {
	s.rg.PreMiddlewares(middlewares...)
}

func (s *Sampan) PostMiddlewares(middlewares ...func(*Context)) {
	s.rg.PostMiddlewares(middlewares...)
}

func (s *Sampan) GetRoute(method string, path string) (handlerChain []func(*Context), params map[string]string) {
	return s.rg.GetRoute(method, path)
}

func (s *Sampan) PutRoute(method string, path string, handler func(*Context)) {
	s.rg.PutRoute(method, path, handler)
}

func (s *Sampan) GET(path string, handler func(*Context)) {
	s.PutRoute(http.MethodGet, path, handler)
}

func (s *Sampan) POST(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPost, path, handler)
}

func (s *Sampan) PUT(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPut, path, handler)
}

func (s *Sampan) PATCH(path string, handler func(*Context)) {
	s.PutRoute(http.MethodPatch, path, handler)
}

func (s *Sampan) DELETE(path string, handler func(*Context)) {
	s.PutRoute(http.MethodDelete, path, handler)
}

func (s *Sampan) HEAD(path string, handler func(*Context)) {
	s.PutRoute(http.MethodHead, path, handler)
}

func (s *Sampan) OPTIONS(path string, handler func(*Context)) {
	s.PutRoute(http.MethodOptions, path, handler)
}

func (s *Sampan) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
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

func (s *Sampan) Listen(addr string) (err error) {
	return http.ListenAndServe(addr, s)
}
