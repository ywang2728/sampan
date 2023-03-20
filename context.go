package sampan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string
	Method     string
	StatusCode int
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    r,
		Path:   r.URL.Path,
		Method: strings.ToUpper(r.Method),
	}
}

func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

func (c *Context) String(code int, format string, value ...any) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	_, err := c.Writer.Write([]byte(fmt.Sprintf(format, value...)))
	if err != nil {
		return
	}
}

func (c *Context) JSON(code int, value any) {
	c.SetHeader("Content-Type", "application/json")
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(value); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

func (c *Context) Data(code int, value []byte) {
	c.Status(code)
	_, err := c.Writer.Write(value)
	if err != nil {
		return
	}
}

func (c *Context) HTML(code int, value string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	_, err := c.Writer.Write([]byte(value))
	if err != nil {
		return
	}
}
