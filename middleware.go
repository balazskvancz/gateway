package gateway

import (
	"sync"
)

type matcherFunc func(*Context) bool

type middleware struct {
	mw      MiddlewareFunc
	matcher matcherFunc
}

type (
	middlewareType string

	middlewareChain []*middleware
)

type middlewareRegistry struct {
	mu       sync.RWMutex
	registry map[middlewareType]middlewareChain
}

const (
	mwPreRunner  middlewareType = "pre"
	mwPostRunner middlewareType = "post"
)

func DefaultMiddlewareMatcher(_ *Context) bool { return true }

// newMiddleware is a factory function for middleware creation.
func newMiddleware(mw MiddlewareFunc, matcher ...matcherFunc) *middleware {
	m := &middleware{
		mw:      mw,
		matcher: DefaultMiddlewareMatcher,
	}

	if len(matcher) > 0 {
		m.matcher = matcher[0]
	}

	return m
}

func (m *middleware) doesMatch(ctx *Context) bool {
	return m.matcher(ctx)
}

// newMiddlewareRegistry creates and returns a new empty middleware registry.
func newMiddlewareRegistry() *middlewareRegistry {
	mw := make(map[middlewareType]middlewareChain)

	mw[mwPreRunner] = make(middlewareChain, 0)
	mw[mwPostRunner] = make(middlewareChain, 0)

	return &middlewareRegistry{
		mu:       sync.RWMutex{},
		registry: mw,
	}
}

// push pushes a middleware into the registry with the given mw type.
func (mr *middlewareRegistry) push(t middlewareType, mw *middleware) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	chain, exists := mr.registry[t]
	if !exists {
		// todo
		return
	}

	mr.registry[t] = append(chain, mw)
}

// get returns the middlewarechain by the given type.
func (mr *middlewareRegistry) get(t middlewareType) middlewareChain {
	chain, exists := mr.registry[t]
	if !exists {
		return nil
	}
	return chain
}

// getHandlerFuncSlice returns a slice of consecutive handlers wrapped inside
// the middleware funcs.
func (chain middlewareChain) getHandlerFuncSlice(next HandlerFunc) []HandlerFunc {
	tlen := len(chain)

	// In case there is not matching middleware, then the only
	// one is the handler itself.
	if tlen == 0 {
		return []HandlerFunc{next}
	}

	funcSlice := make([]HandlerFunc, tlen)

	funcSlice[tlen-1] = func(ctx *Context) {
		chain[tlen-1].mw(ctx, next)
	}

	for i := tlen - 2; i > 0; i-- {
		mw := chain[i].mw

		funcSlice[i] = func(ctx *Context) {
			mw(ctx, funcSlice[i+1])
		}
	}

	return funcSlice
}
