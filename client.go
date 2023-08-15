package gateway

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 3 sec
	defaultClientTimeout time.Duration = 3000 * time.Millisecond
)

type client struct {
	// ctx context.Context
	*http.Client

	//
	hostName string
}

type httpClient interface {
	doRequest(string, string, io.Reader, ...http.Header) (*http.Response, error)
	pipe(*http.Request) (*http.Response, error)
	Do(*http.Request) (*http.Response, error)
}

type httpClientOptionFunc func(*client)

func withHostName(hname string) httpClientOptionFunc {
	return func(hc *client) {
		hc.hostName = hname
	}
}

// newHttpClient returns a new client.
func newHttpClient(opts ...httpClientOptionFunc) httpClient {
	hc := &client{
		Client: &http.Client{
			Timeout: defaultClientTimeout,
		},
	}

	for _, o := range opts {
		o(hc)
	}

	return hc
}

type reqConfig struct {
	method string
	url    string
	header http.Header
	body   io.Reader
}

func (cl *client) doRequest(method string, url string, body io.Reader, header ...http.Header) (*http.Response, error) {
	finalHeader := func() http.Header {
		if len(header) > 0 {
			return header[0]
		}
		return http.Header{}
	}()

	return cl.do(reqConfig{
		method: method,
		url:    cl.hostName + url,
		body:   body,
		header: finalHeader,
	})
}

func (cl *client) pipe(req *http.Request) (*http.Response, error) {
	body := req.Body
	defer body.Close()

	return cl.do(reqConfig{
		method: req.Method,
		url:    cl.hostName + req.RequestURI,
		header: req.Header,
		body:   body,
	})
}

func (cl *client) do(conf reqConfig) (*http.Response, error) {
	req, err := http.NewRequest(conf.method, conf.url, conf.body)
	if err != nil {
		return nil, err
	}

	// If there is customHeader, then we add the missing ones.
	if conf.header != nil {
		for k, v := range conf.header {
			req.Header.Add(k, strings.Join(v, "; "))
		}
	}

	// Always close each request after it is done.
	req.Close = true

	return cl.Do(req)
}
