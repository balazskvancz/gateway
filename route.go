package gateway

import (
	"strings"
)

var (
	query = '?'
)

// removeQueryParts removes the query strings from
// the given url, if there is any.
func removeQueryParts(url string) string {
	idx := strings.IndexRune(url, query)
	if idx > 0 {
		return url[:idx]
	}
	return url
}

type Route struct {
	fullUrl string

	chain []HandlerFunc
}

func newRoute(url string, fn HandlerFunc) *Route {
	return &Route{
		fullUrl: url,
		chain:   []HandlerFunc{fn},
	}
}

func (route *Route) registerMiddleware(mw MiddlewareFunc) *Route {
	if len(route.chain) == 0 {
		return route
	}

	chain := route.chain

	var mwFun HandlerFunc = func(ctx *Context) {
		mw(ctx, chain[0])
	}

	route.chain = append([]HandlerFunc{mwFun}, route.chain...)

	return route
}

// RegisterMiddlewares registers all the given middlewares one-by-one,
// then returns the route pointer.
func (route *Route) RegisterMiddlewares(mws ...MiddlewareFunc) *Route {
	if len(mws) == 0 {
		return route
	}

	// Have to register in reversed order.
	for i := len(mws) - 1; i >= 0; i-- {
		route.registerMiddleware(mws[i])
	}

	return route
}

func (route *Route) getChain() HandlerFunc {
	return route.chain[0]
}

// getHandler returns the actual handler, which is at the end of the chain.
func (route *Route) getHandler() HandlerFunc {
	return route.chain[len(route.chain)-1]
}

func getParamKey(val string) string {
	if !strings.HasPrefix(val, string(curlyStart)) && !strings.HasSuffix(val, string(curlyEnd)) {
		return ""
	}
	return val[1 : len(val)-1]
}

func getPathParams(storedUrl, incomingUrl string) []pathParam {
	count := strings.Count(storedUrl, string(curlyStart))

	params := make([]pathParam, count)

	var (
		// The length of these two MUST be equal.
		storedSplitted   = strings.Split(storedUrl, string(slash))
		incomingSplitted = strings.Split(incomingUrl, string(slash))
		counter          = 0
	)

	// However, we trust no one, so one last check.
	if len(storedSplitted) != len(incomingSplitted) {
		return params
	}

	for i := 0; i < len(storedSplitted); i++ {
		key := getParamKey(storedSplitted[i])

		if key != "" {
			param := pathParam{
				key:   key,
				value: incomingSplitted[i],
			}

			params[counter] = param
			counter++
		}
	}

	return params
}
