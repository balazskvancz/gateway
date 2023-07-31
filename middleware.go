package gateway

type MiddlewareType string

const (
	MiddlewarePreRunner  MiddlewareType = "preRunner"
	MiddlewarePostRunner MiddlewareType = "postRunner"
)

type Middleware interface {
	DoesMatch(*Context) bool
	IsPreRunner() bool
	Exec(*Context, HandlerFunc)
}

type MatcherFunc func(*Context) bool

type middleware struct {
	mw      MiddlewareFunc
	matcher MatcherFunc
	t       MiddlewareType
}

var _ Middleware = (*middleware)(nil)

type (
	middlewareChain []Middleware
)

type matchingMiddleware struct {
	pre  middlewareChain
	post middlewareChain
}

func DefaultMiddlewareMatcher(_ *Context) bool { return true }

type middlewareOptionFunc func(*middleware)

func withMiddlewareType(t MiddlewareType) middlewareOptionFunc {
	return func(m *middleware) {
		m.t = t
	}
}

func withMiddlewareMatcherFunc(fn MatcherFunc) middlewareOptionFunc {
	return func(m *middleware) {
		m.matcher = fn
	}
}

// newMiddleware is a factory function for middleware creation.
func newMiddleware(mwFunc MiddlewareFunc, opts ...middlewareOptionFunc) Middleware {
	mw := &middleware{
		mw:      mwFunc,
		matcher: DefaultMiddlewareMatcher,
	}

	for _, o := range opts {
		o(mw)
	}

	return mw
}

// DoesMatch returns wether a MW mathes for a given Context or not.
func (m *middleware) DoesMatch(ctx *Context) bool {
	return m.matcher(ctx)
}

// IsPreRunner returns wether a MW is prerunner or not.
func (m *middleware) IsPreRunner() bool {
	return m.t == MiddlewarePreRunner
}

// Exec executes a given Middleware with given
// Context and HandlerFunc to call as next.
func (m *middleware) Exec(ctx *Context, next HandlerFunc) {
	m.mw(ctx, next)
}

// createChain creates and returns a chain of handlerFuncs from original
// middlewareChain, with the last element as the given handlerFunc.
func (chain middlewareChain) createChain(next HandlerFunc) []HandlerFunc {
	tlen := len(chain)

	if tlen == 0 {
		return []HandlerFunc{next}
	}

	funcSlice := make([]HandlerFunc, tlen)

	funcSlice[tlen-1] = func(ctx *Context) {
		chain[tlen-1].Exec(ctx, next)
	}

	for i := tlen - 2; i >= 0; i-- {
		var idx = i
		funcSlice[i] = func(ctx *Context) {
			chain[idx].Exec(ctx, funcSlice[idx+1])
		}
	}

	return funcSlice
}

// getHandler returns the tied middleware functions from the matching pre
// and post middlewares with the actual handlerfunc.
func (mm *matchingMiddleware) getHandler(handler HandlerFunc) HandlerFunc {
	postChain := mm.post.createChain(writeToResponseMiddleware)

	// Wrap the given HandlerFunc inside a MW func, which calls
	// the first element of the post chain.
	handlerMw := func(ctx *Context) {
		handler(ctx)
		if len(postChain) > 0 {
			postChain[0](ctx)
		}
	}

	preChain := mm.pre.createChain(handlerMw)

	return preChain[0]
}
