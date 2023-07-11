package gateway

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/balazskvancz/gateway/pkg/utils"
)

const (
	ctHeader = "Content-Type"

	JsonContentType     = "application/json"
	JsonContentTypeUTF8 = JsonContentType + "; charset=UTF-8"
	TextHtmlContentType = "text/html"
	XmlContentType      = "application/xml"

	maxParams uint8 = 8
)

var (
	defaultNotFoundBody []byte = []byte("404 – Not Found")
)

type pathParam struct {
	key   string
	value string
}

type contextIdChan <-chan uint64

type Context struct {
	writer  http.ResponseWriter
	request *http.Request

	params []pathParam

	contextId     uint64
	contextIdChan contextIdChan
}

// newContext creates and returns a new context.
func newContext(ciChan contextIdChan) *Context {
	return &Context{
		contextIdChan: ciChan,
		params:        make([]pathParam, maxParams),
	}
}

// reset
func (ctx *Context) reset(w http.ResponseWriter, r *http.Request) {
	ctx.request = r
	ctx.writer = w
	ctx.params = ctx.params[:0]

	// És kiolvassuk az channelből érkező azonosítót is.
	ctx.contextId = <-ctx.contextIdChan
}

// empty
func (c *Context) empty() {
	c.request = nil
	c.writer = nil
}

// Request.
// Returns the attached request.
func (ctx *Context) GetRequest() *http.Request {
	return ctx.request
}

// Returns the method of incoming request.
func (ctx *Context) GetRequestMethod() string {
	if ctx == nil {
		return ""
	}
	return ctx.request.Method
}

// Returns the full URL with all queryParams included.
func (ctx *Context) GetFullUrl() string {
	if ctx.request == nil {
		return ""
	}
	return ctx.request.RequestURI
}

// Return the array, of url parts.
// /foo/bar/baz => ["foo", "bar", "baz"]
func (ctx *Context) GetUrlParts() []string {
	return utils.GetUrlParts(ctx.GetFullUrl())
}

// Returns the url, without query params, it there is any.
func (ctx *Context) GetUrlWithoutQueryParams() string {
	return removeQueryParts(ctx.GetFullUrl())
}

func (ctx *Context) GetQueryParams() url.Values {
	return ctx.request.URL.Query()
}

// Returns one params value based on its key.
func (ctx *Context) GetQueryParam(key string) string {
	query := ctx.GetQueryParams()

	return query.Get(key)
}

// Get the body in bytes.
func (ctx *Context) GetRawBody() ([]byte, error) {
	req := ctx.request

	b, err := io.ReadAll(req.Body)

	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	return b, nil
}

// Reads the request body, tries to parse into given object.
// It must be a pointer, otherwise wont work.
// Also, it returns error, if somethting went bad.
func (ctx *Context) ReadJsonBody(data interface{}) error {
	b, err := ctx.GetRawBody()

	if err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}

// Returns all the headers from the request.
func (ctx *Context) GetRequestHeaders() http.Header {
	return ctx.request.Header
}

// Return one specific headers value, with given key.
func (ctx *Context) GetRequestHeader(key string) string {
	header := ctx.GetRequestHeaders()

	return header.Get(key)
}

// Returns te content-type of the original request.
func (ctx *Context) GetContentType() string {
	return ctx.GetRequestHeader(ctHeader)
}

// ----------------
// | ROUTE PARAMS |
// ----------------

// Sets the params to the context.
func (ctx *Context) setParams(params []pathParam) {
	ctx.params = params
}

// Get params value by certain key.
func (ctx *Context) GetParam(key string) string {
	for _, entry := range ctx.params {
		if entry.key == key {
			return entry.value
		}
	}

	return ""
}

// Response

// Writes the response body, with given byte[] and Content-type.
func (ctx *Context) SendRaw(b []byte, statusCode int, header http.Header) {
	writer := ctx.writer

	for k, v := range header {
		var value string

		for i, vv := range v {
			value = value + vv

			if i != len(v)-1 {
				value += ", "
			}
		}

		writer.Header().Add(k, strings.Join(v, ";"))
	}

	writer.WriteHeader(statusCode)
	writer.Write(b)
}

// Sends JSON response to client.
func (ctx *Context) SendJson(data interface{}) {
	b, err := json.Marshal(data)

	if err != nil {
		fmt.Printf("marshal err: %v\n", err)

		return
	}

	ctx.SendRaw(b, http.StatusOK, createContentTypeHeader(JsonContentType))
}

// Creates a content type header.
func createContentTypeHeader(ct string) http.Header {
	header := http.Header{}

	header.Add("Content-Type", ct)

	return header
}

// Send XML response to client.
func (ctx *Context) SendXML(data interface{}) {
	b, err := xml.Marshal(data)

	if err != nil {
		fmt.Printf("marshal err: %v\n", err)

		return
	}

	ctx.SendRaw(b, http.StatusOK, createContentTypeHeader(XmlContentType))
}

// Sending a HTTP 404 error.
func (ctx *Context) SendNotFound() {
	ctx.writer.WriteHeader(http.StatusNotFound)
	ctx.writer.Write(defaultNotFoundBody)
}

func (ctx *Context) SendInternalServerError() {
	ctx.writer.WriteHeader(http.StatusInternalServerError)
}

// Sending basic HTTP 200.
func (ctx *Context) SendOk() {
	ctx.SendRaw(nil, http.StatusOK, http.Header{})
}

// Sending HTTP 401 error, if the request
// doesnt have the required permissions.
func (ctx *Context) SendUnauthorized() {
	ctx.SendRaw(nil, http.StatusUnauthorized, http.Header{})
}

// Sends `service unavailable` error, if the
// given service hasnt responded or its state
// is also unavailable.
func (ctx *Context) SendUnavailable() {
	ctx.SendRaw(nil, http.StatusServiceUnavailable, http.Header{})
}

// Sends a text error with HTTP 400 code in header.
func (ctx *Context) SendError(msg ...string) {
	b := []byte{}

	if len(msg) > 0 {
		b = []byte(msg[0])
	}

	ctx.SendRaw(b, http.StatusBadRequest, createContentTypeHeader(TextHtmlContentType))
}
