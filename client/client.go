package client

import (
	"net/http"
	"time"
)

const (
	DEFAULT_TIMEOUT = 10 * time.Second
)

type (
	Request struct {
		req     *http.Request
		header  *http.Header
		Cookies []*http.Cookie
	}

	Response struct {
		resp   *http.Response
		Header *http.Header
		Body   []byte
	}

	HttpClient struct {
	}
)

func Get(url string, headers ...http.Header) (*Response, error) {

}

func Post(url string, body any, headers ...http.Header) (*Response, error) {

}
func Put(url string, body any, headers ...http.Header) (*Response, error) {

}
func Patch(url string, body any, headers ...http.Header) (*Response, error) {

}
func Delete(url string, headers ...http.Header) (*Response, error) {

}
func Options(url string, headers ...http.Header) (*Response, error) {

}
