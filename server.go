package sampan

type Server struct {
	router *router
}

func New() *Server {
	return &Server{router: newRouter()}
}

/*func (s *Server) AddRoute(method string, path string, handler func(*Context)) {
	s.router.put(method, path, handler)
}

func (s *Server) GET(path string, handler func(*Context)) {
	s.AddRoute(http.MethodGet, path, handler)
}

func (s *Server) POST(path string, handler func(*Context)) {
	s.AddRoute(http.MethodPost, path, handler)
}

func (s *Server) PUT(path string, handler func(*Context)) {
	s.AddRoute(http.MethodPut, path, handler)
}

func (s *Server) DELETE(path string, handler func(*Context)) {
	s.AddRoute(http.MethodDelete, path, handler)
}

func (s *Server) handle(c *Context) {
	log.Printf("Handle %4s - %s", c.Method, c.Path)
	handler := s.router.get(c.Method, c.Path)
	if handler != nil {
		handler(c)
	} else {
		c.String(http.StatusNotFound, "404 NOT FOUND: %s\n", c.Path)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	s.handle(newContext(w, req))
}

func (s *Server) Run(addr string) (err error) {
	return http.ListenAndServe(addr, s)
}*/
