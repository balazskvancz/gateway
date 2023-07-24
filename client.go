package gateway

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// 3 sec
	defaultClientTimeout time.Duration = 3000 * time.Millisecond
)

type httpClient struct {
	ctx context.Context
	*http.Client

	//
	hostName string
}

type httpClientOptionFunc func(*httpClient)

func withHostName(hname string) httpClientOptionFunc {
	return func(hc *httpClient) {
		hc.hostName = hname
	}
}

// newHttpClient returns a new client.
func newHttpClient(opts ...httpClientOptionFunc) *httpClient {
	hc := &httpClient{
		Client: &http.Client{
			Timeout: defaultClientTimeout,
		},
		ctx: context.Background(),
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

func (cl *httpClient) doRequest(method string, url string, body io.Reader, header ...http.Header) (*http.Response, error) {
	finalHeader := func() http.Header {
		if len(header) > 0 {
			return header[0]
		}
		return http.Header{}
	}()

	return cl.doWithTimeout(reqConfig{
		method: method,
		url:    cl.hostName + url,
		body:   body,
		header: finalHeader,
	})
}

func (cl *httpClient) pipe(req *http.Request) (*http.Response, error) {
	body := req.Body
	defer body.Close()

	return cl.doWithTimeout(reqConfig{
		method: req.Method,
		url:    cl.hostName + req.RequestURI,
		header: req.Header,
		body:   body,
	})
}

func (cl *httpClient) doWithTimeout(conf reqConfig) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(cl.ctx, cl.Timeout)
	defer cancel()

	return cl.do(ctx, conf)
}

func (cl *httpClient) do(ctx context.Context, conf reqConfig) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, conf.method, conf.url, conf.body)
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
