package gateway

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
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

// reset resets the context entity to default state.
func (ctx *Context) reset(w http.ResponseWriter, r *http.Request) {
	ctx.request = r
	ctx.writer = w
	ctx.params = ctx.params[:0]

	// És kiolvassuk az channelből érkező azonosítót is.
	ctx.contextId = <-ctx.contextIdChan
}

// empty makes the http.Request and http.ResponseWrite <nil>.
// Should be called before putting the Context back to the pool.
func (c *Context) empty() {
	c.request = nil
	c.writer = nil
}

// GetRequest returns the attached http.Request.
func (ctx *Context) GetRequest() *http.Request {
	return ctx.request
}

// GetRequestMethod returns the method of incoming request.
func (ctx *Context) GetRequestMethod() string {
	if ctx == nil {
		return ""
	}
	return ctx.request.Method
}

// GetFullUrl returns the full URL with all queryParams included.
func (ctx *Context) GetFullUrl() string {
	if ctx.request == nil {
		return ""
	}
	return ctx.request.RequestURI
}

// GetUrlParts returns the url as a slice of strings
func (ctx *Context) GetUrlParts() []string {
	return utils.GetUrlParts(ctx.GetFullUrl())
}

// GetUrlWithoutQueryParams returns the url
// without query params, it there is any.
func (ctx *Context) GetUrlWithoutQueryParams() string {
	return removeQueryParts(ctx.GetFullUrl())
}

// GetQueryParams returns the query params of the url.
func (ctx *Context) GetQueryParams() url.Values {
	return ctx.request.URL.Query()
}

// GetQueryParam returns the queryParam identified by the given key.
func (ctx *Context) GetQueryParam(key string) string {
	query := ctx.GetQueryParams()

	return query.Get(key)
}

// GetRawBody reads and returns the body of the request.
func (ctx *Context) GetRawBody() ([]byte, error) {
	req := ctx.request

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	defer req.Body.Close()

	return b, nil
}

// DecodeJsonBody decodes the body into the given paramer.
// The given parameter must be a pointer type, otwherwise
// it returns an error.
func (ctx *Context) DecodeJsonBody(data interface{}) error {
	if ct := ctx.GetContentType(); !strings.Contains(ct, JsonContentType) {
		return errNotJsonContentType
	}

	if reflect.ValueOf(data).Kind() != reflect.Ptr {
		return errDataMustBePtr
	}
	body := ctx.GetRequest().Body
	defer body.Close()
	return json.NewDecoder(body).Decode(data)
}

// GetRequestHeaders returns all the headers from the request.
func (ctx *Context) GetRequestHeaders() http.Header {
	return ctx.request.Header
}

// GetRequestHeader return one specific headers value, with given key.
func (ctx *Context) GetRequestHeader(key string) string {
	header := ctx.GetRequestHeaders()

	return header.Get(key)
}

// GetContentType returns te content-type of the original request.
func (ctx *Context) GetContentType() string {
	return ctx.GetRequestHeader(ctHeader)
}

// ----------------
// | ROUTE PARAMS |
// ----------------

// setParams binds the params to the context.
func (ctx *Context) setParams(params []pathParam) {
	ctx.params = params
}

// GetParam returns the value of the param identified by the given key.
func (ctx *Context) GetParam(key string) string {
	for _, entry := range ctx.params {
		if entry.key == key {
			return entry.value
		}
	}

	return ""
}

// Writes the response body, with given byte[] and Content-type.
func (ctx *Context) SendRaw(b []byte, statusCode int, header http.Header) {
	writer := ctx.writer

	writer.WriteHeader(statusCode)
	writer.Write(b)
	ctx.appendHttpHeader(header)
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

// SendNotFound sends a HTTP 404 error.
func (ctx *Context) SendNotFound() {
	ctx.SendHttpError(http.StatusNotFound)
}

// SendInternalServerError send a HTTP 500 error.
func (ctx *Context) SendInternalServerError() {
	ctx.SendHttpError(http.StatusInternalServerError)
}

// SendOk send a s basic HTTP 200 response.
func (ctx *Context) SendOk() {
	ctx.SendRaw(nil, http.StatusOK, http.Header{})
}

// SendUnauthorized send a HTTP 401 error.
func (ctx *Context) SendUnauthorized() {
	ctx.SendHttpError(http.StatusUnauthorized)
}

// SendUnavailable send a HTTP 503 error.
func (ctx *Context) SendUnavailable() {
	ctx.SendHttpError(http.StatusServiceUnavailable)
}

// SendHttpError send HTTP error with the given code.
// It also write the statusText inside the body, based on the code.
func (ctx *Context) SendHttpError(code int) {
	http.Error(ctx.writer, http.StatusText(code), code)
}

// SendError sends a text error with HTTP 400 code in header.
func (ctx *Context) SendError(msg ...string) {
	b := []byte{}

	if len(msg) > 0 {
		b = []byte(msg[0])
	}

	ctx.SendRaw(b, http.StatusBadRequest, createContentTypeHeader(TextHtmlContentType))
}

// Pipe writes the given repsonse's body, statusCode and headers to the Context's response.
func (ctx *Context) Pipe(res *http.Response) {
	// We could use TeeReader if we want to know
	// what are we writing to the request.
	// r := io.TeeReader(res.Body, ctx.writer)
	io.Copy(ctx.writer, res.Body)
	ctx.appendHttpHeader(res.Header)
	ctx.writer.WriteHeader(res.StatusCode)
}

func (ctx *Context) appendHttpHeader(header http.Header) {
	for k, v := range header {
		ctx.writer.Header().Add(k, strings.Join(v, "; "))
	}
}
