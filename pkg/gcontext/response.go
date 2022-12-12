package gcontext

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Writes the response body, with given byte[] and Content-type.
func (ctx *GContext) SendRaw(b []byte, statusCode int, header http.Header) {
	ctx.nextCalled = false

	writer := ctx.writer

	var (
		cType string
	)

	for k, v := range header {
		var value string

		for i, vv := range v {
			value = value + vv

			if i != len(v)-1 {
				value += ", "
			}
		}

		if strings.ToLower(k) == "content-type" {
			cType = value
		}

		writer.Header().Add(k, strings.Join(v, ";"))
	}

	action := strconv.Itoa(statusCode) + " " + cType
	go ctx.logger.writeToLog(ctx.contextId, stateEnded, action)

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
