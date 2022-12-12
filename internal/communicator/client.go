package communicator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

type HttpClient struct {
	*http.Client

	chansPool sync.Pool

	header http.Header

	hostPort   string // Contains the host:port of the given service.
	timeOutSec int    // Timeout, for each HTTP request.
}

type GoRequest interface {
	GoRequest(method string, url string, data []byte, chans *concurrentChan)
}

// Make sure, that is satisfies the interface.
var _ GoRequest = (*HttpClient)(nil)

// Creates a new pointer.
func New(hostPort string, timeOutSec int) *HttpClient {
	return &HttpClient{
		chansPool: sync.Pool{
			New: func() interface{} { return new(GChan) },
		},
		Client:     &http.Client{},
		hostPort:   hostPort,
		timeOutSec: timeOutSec,
	}
}

// Return a pointer to GChan from the pool.
func (client *HttpClient) PoolGet() *GChan {
	return client.chansPool.Get().(*GChan)
}

// Puts a GChan pointer to the pool.
func (client *HttpClient) PoolPut(ch *GChan) {
	client.chansPool.Put(ch)
}

// Setting the HTTP Header for the client.
func (client *HttpClient) SetHeader(h http.Header) {
	client.header = h
}

// Basic forwarder, based on the request.
// Method safe with concurrency.
func (client *HttpClient) Forwarder(req *http.Request) ([]byte, int, http.Header) {
	client.header = req.Header
	chans := client.chansPool.Get().(*GChan)

	defer client.chansPool.Put(chans)

	method, url := req.Method, req.RequestURI

	if method == http.MethodGet {
		res, err := client.Get(url)

		if err != nil {
			fmt.Printf("[FORWARDER]: %v\n", err)

			return []byte{}, http.StatusInternalServerError, http.Header{}
		}
		defer res.Body.Close()

		return parseResponse(res)
	}

	if method == http.MethodPost {
		body, err := io.ReadAll(req.Body)

		// In case of bad body.
		if err != nil {
			return nil, http.StatusBadRequest, nil
		}

		res, err := client.Post(url, body)

		if err != nil {
			fmt.Printf("[FORWARDER]: %v\n", err)

			return []byte{}, http.StatusInternalServerError, http.Header{}
		}
		defer res.Body.Close()

		return parseResponse(res)
	}

	return []byte{}, http.StatusBadRequest, http.Header{}
}

// Get request.
func (client *HttpClient) Get(url string) (*http.Response, error) {
	return client.request(http.MethodGet, url, nil)
}

// Post method.
func (client *HttpClient) Post(url string, data []byte) (*http.Response, error) {
	return client.request(http.MethodPost, url, data)
}

// Goroutine supporting request function.
func (client *HttpClient) GoRequest(method, url string, data []byte, chans *concurrentChan) {
	res, err := client.request(method, url, data)

	chans.url <- url

	if err != nil {
		chans.GChan.Data <- nil
		chans.GChan.Header <- nil
		chans.GChan.Status <- http.StatusInternalServerError

		return
	}

	b, err := io.ReadAll(res.Body)

	if err != nil {
		chans.GChan.Data <- nil
		chans.GChan.Header <- nil
		chans.GChan.Status <- http.StatusInternalServerError

		return
	}

	chans.GChan.Data <- b
	chans.GChan.Header <- res.Header
	chans.GChan.Status <- res.StatusCode

	res.Body.Close()
}

// The abstraction above the native HTTP request by the client.
func (client *HttpClient) request(method, url string, data []byte) (*http.Response, error) {
	// ctx, cancel := context.WithTimeout(context.Background(), client.Timeout * time.Second)

	var body io.Reader = nil

	// If there is any data given, append to the body.
	if data != nil {
		body = bytes.NewReader(data)
	}

	if !strings.HasPrefix(url, client.hostPort) {
		url = client.hostPort + url
	}

	r, err := http.NewRequest(method, url, body)

	if err != nil {
		fmt.Printf("%v\n", err)

		return nil, err
	}

	ct := r.Header.Get("Content-Type")
	cl := r.Header.Get("Content-Length")

	// Set the header.
	r.Header = client.header

	if ct != "" {
		r.Header.Add("Content-Type", ct)
	}

	if cl != "" {
		r.Header.Add("Content-Lenght", cl)
	}

	res, err := client.Do(r)

	if err != nil {
		fmt.Printf("%v\n", err)

		return nil, err
	}

	return res, nil
}

func parseResponse(res *http.Response) ([]byte, int, http.Header) {
	body, err := io.ReadAll(res.Body)

	if err != nil {
		return []byte{}, http.StatusInternalServerError, nil
	}

	return body, res.StatusCode, res.Header
}
