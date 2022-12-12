package gcontext

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
)

const (
	ctHeader = "Content-Type"

	JsonContentType     = "application/json"
	TextHtmlContentType = "text/html"
	XmlContentType      = "application/xml"

	queryParamSeparator = "&"
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

	logger *logger
}

// Creates and returns a new pointer to gctx.
func New(w http.ResponseWriter, r *http.Request) *GContext {
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
	if ctx.logger == nil {
		ctx.logger = NewLogger()
	}

	ctx.request = r
	ctx.writer = w
	ctx.mwIndex = 0
	ctx.nextCalled = false
	ctx.contextId = contextId

	ctx.mutex.Unlock()

	// Int this case, there is a new request
	// so we should write it to the log file.
	action := fmt.Sprintf("%s %s", ctx.GetRequestMethod(), ctx.GetFullUrl())
	go ctx.logger.writeToLog(contextId, stateStarted, action)
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
