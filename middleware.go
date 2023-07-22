package gateway

type MiddlewareType string

const (
	MiddlewarePreRunner MiddlewareType = "preRunner"
	MiddlwarePostRunner MiddlewareType = "postRunner"
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

func (m *middleware) DoesMatch(ctx *Context) bool {
	return m.matcher(ctx)
}

func (m *middleware) IsPreRunner() bool {
	return m.t == MiddlewarePreRunner
}

func (m *middleware) Exec(ctx *Context, next HandlerFunc) {
	m.mw(ctx, next)
}

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
