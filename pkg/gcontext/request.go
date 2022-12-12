package gcontext

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/balazskvancz/gateway/pkg/utils"
)

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
