package gateway

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/balazskvancz/gateway/pkg/utils"
)

const (
	ctHeader = "Content-Type"

	JsonContentType     = "application/json"
	TextHtmlContentType = "text/html"
	XmlContentType      = "application/xml"
)

var (
	errParamNotExists = errors.New("requested param not exists")

	contextId uint64 = 0
)

type GContext struct {
	writer  http.ResponseWriter
	request *http.Request

	params map[string]string

	mwIndex    uint8
	nextCalled bool

	contextId uint64

	mutex sync.Mutex

	// logger *logger
}

// Creates and returns a new pointer to gctx.
func newContext(w http.ResponseWriter, r *http.Request) *GContext {
	// logger := NewLogger()

	return &GContext{
		writer:    w,
		request:   r,
		mwIndex:   0,
		contextId: 0,
		mutex:     sync.Mutex{},
	}
}

// Sets context to default state.
func (ctx *GContext) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.mutex.Lock()
	contextId += 1

	// Just in case, if the logger is nil, we should create it.
	// if ctx.logger == nil {
	// ctx.logger = NewLogger()
	// }

	ctx.request = r
	ctx.writer = w
	ctx.mwIndex = 0
	ctx.nextCalled = false
	ctx.contextId = contextId

	ctx.mutex.Unlock()

	// Int this case, there is a new request
	// so we should write it to the log file.
	// action := fmt.Sprintf("%s %s", ctx.GetRequestMethod(), ctx.GetFullUrl())
	// go ctx.logger.writeToLog(contextId, stateStarted, action)
}

// ------------------
// | 	 MIDDLEWARE   |
// ------------------

// Calls the next element in handlerChain.
func (ctx *GContext) Next() {
	ctx.nextCalled = true
}

// Increments the middlewareIndex and resets the "nextCalled" flag.
func (ctx *GContext) NextIteration() {
	ctx.mwIndex = ctx.mwIndex + 1
	ctx.nextCalled = false
}

// Returns whether the next is called in the chain.
func (ctx *GContext) IsNextCalled() bool {
	return ctx.nextCalled
}

// Returns the current index.
func (ctx *GContext) GetIndex() uint8 {
	return ctx.mwIndex
}

// Request.

// Returns the attached request.
func (ctx *GContext) GetRequest() *http.Request {
	return ctx.request
}

// Returns the method of incoming request.
func (ctx *GContext) GetRequestMethod() string {
	if ctx == nil {
		return ""
	}

	return ctx.request.Method
}

// Returns the full URL with all queryParams included.
func (ctx *GContext) GetFullUrl() string {
	if ctx.request == nil {
		fmt.Println("[CONTEXT]: request is nil")
		return ""
	}

	return ctx.request.RequestURI
}

// Return the array, of url parts.
// /foo/bar/baz => ["foo", "bar", "baz"]
func (ctx *GContext) GetUrlParts() []string {
	fullUrl := ctx.GetFullUrl()

	return utils.GetUrlParts(fullUrl)
}

// Returns the url, without query params, it there is any.
func (ctx *GContext) GetUrlWithoutQueryParams() string {
	fullUrl := ctx.GetFullUrl()

	index := strings.Index(fullUrl, "?")

	if index > 0 {
		return fullUrl[:index]
	}

	return fullUrl
}

func (ctx *GContext) GetQueryParams() *map[string]string {
	params, url := make(map[string]string), ctx.GetFullUrl()

	index := strings.Index(url, "?")

	// If there no "?", there isnt any param.
	if index == -1 {
		return &params
	}

	// Remove the first of the part of the url.
	paramsPart := url[(index + 1):]

	// Every param is divided by "&".
	allParams := strings.Split(paramsPart, "&")

	for _, val := range allParams {
		splitted := strings.Split(val, "=")

		if len(splitted) > 1 {
			params[splitted[0]] = splitted[1]
		}
	}

	return &params
}

// Returns one params value based on its key.
func (ctx *GContext) GetQueryParam(key string) string {
	params := ctx.GetQueryParams()

	v, e := (*params)[key]

	if !e {
		return ""
	}

	return v
}

// Get the body in bytes.
func (ctx *GContext) GetRawBody() ([]byte, error) {
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
func (ctx *GContext) ReadJsonBody(data interface{}) error {
	b, err := ctx.GetRawBody()

	if err != nil {
		return err
	}

	return json.Unmarshal(b, data)
}

// Returns all the headers from the request.
func (ctx *GContext) GetRequestHeaders() http.Header {
	return ctx.request.Header
}

// Return one specific headers value, with given key.
func (ctx *GContext) GetRequestHeader(key string) string {
	header := ctx.GetRequestHeaders()

	return header.Get(key)
}

// Returns te content-type of the original request.
func (ctx *GContext) GetContentType() string {
	return ctx.GetRequestHeader(ctHeader)
}

// ----------------
// | ROUTE PARAMS |
// ----------------

// Sets the params to the context.
func (ctx *GContext) SetParams(p map[string]string) {
	ctx.params = p
}

// Get params value by certain key.
func (ctx *GContext) GetParam(key string) (string, error) {
	v, ex := ctx.params[key]

	if !ex {
		return "", errParamNotExists
	}

	return v, nil
}

// Response

// Writes the response body, with given byte[] and Content-type.
func (ctx *GContext) SendRaw(b []byte, statusCode int, header http.Header) {
	ctx.nextCalled = false

	writer := ctx.writer

	for k, v := range header {
		var value string

		for i, vv := range v {
			value = value + vv

			if i != len(v)-1 {
				value += ", "
			}
		}

		// if strings.ToLower(k) == "content-type" {
		// cType = value
		// }

		writer.Header().Add(k, strings.Join(v, ";"))
	}

	// action := strconv.Itoa(statusCode) + " " + cType
	// go ctx.logger.writeToLog(ctx.contextId, stateEnded, action)

	writer.WriteHeader(statusCode)
	writer.Write(b)
}

// Sends JSON response to client.
func (ctx *GContext) SendJson(data interface{}) {
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
func (ctx *GContext) SendXML(data interface{}) {
	b, err := xml.Marshal(data)

	if err != nil {
		fmt.Printf("marshal err: %v\n", err)

		return
	}

	ctx.SendRaw(b, http.StatusOK, createContentTypeHeader(XmlContentType))
}

// Sending a HTTP 404 error.
func (ctx *GContext) SendNotFound() {
	ctx.nextCalled = false
	ctx.writer.WriteHeader(http.StatusNotFound)
}

// Sending basic HTTP 200.
func (ctx *GContext) SendOk() {
	ctx.nextCalled = false
	ctx.SendRaw(nil, http.StatusOK, http.Header{})
}

// Sending HTTP 401 error, if the request
// doesnt have the required permissions.
func (ctx *GContext) SendUnauthorized() {
	ctx.nextCalled = false
	ctx.SendRaw(nil, http.StatusUnauthorized, http.Header{})
}

// Sends `service unavailable` error, if the
// given service hasnt responded or its state
// is also unavailable.
func (ctx *GContext) SendUnavailable() {
	ctx.nextCalled = false
	ctx.SendRaw(nil, http.StatusServiceUnavailable, http.Header{})
}

// Sends a text error with HTTP 400 code in header.
func (ctx *GContext) SendError(msg ...string) {
	ctx.nextCalled = false
	b := []byte{}

	if len(msg) > 0 {
		b = []byte(msg[0])
	}

	ctx.SendRaw(b, http.StatusBadRequest, createContentTypeHeader(TextHtmlContentType))
}
