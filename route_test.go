package gateway

import (
	"reflect"
	"testing"
)

type routeFactoryFn func(t *testing.T) *Route

func TestRegisterMiddleware(t *testing.T) {
	var testW []byte = []byte("written")

	var testHandler = func(ctx *Context) {
		ctx.writer.write(testW)
	}

	type testCase struct {
		name       string
		getRoute   routeFactoryFn
		mwFunction []MiddlewareFunc
	}

	tt := []testCase{
		{
			name: "test for zero middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{},
		},
		{
			name: "test for one middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{func(ctx *Context, _ HandlerFunc) {
				ctx.writer.write([]byte("first-mw"))
			}},
		},
		{
			name: "test for two middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{func(ctx *Context, _ HandlerFunc) {
				ctx.writer.write([]byte("first-mw"))
			},
				func(ctx *Context, _ HandlerFunc) {
					ctx.writer.write([]byte("second"))
				}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			route := tc.getRoute(t)

			ctx := newContext(nil, nil)

			for _, mw := range tc.mwFunction {
				route.registerMiddleware(mw)
			}

			var (
				expLen = len(tc.mwFunction) + 1
				gotLen = len(route.chain)
			)

			if gotLen != expLen {
				t.Errorf("expected chain length: %d; got length: %d\n", expLen, gotLen)
			}

			route.getHandler()(ctx)

			if !reflect.DeepEqual(ctx.writer.b, testW) {
				t.Errorf("expected writte value: %s; actually written value: %s\n", string(testW), string(ctx.writer.b))
			}
		})
	}

}

func TestRegisterMiddlewares(t *testing.T) {
	var testW []byte = []byte("written")

	var testHandler = func(ctx *Context) {
		ctx.writer.write(testW)
	}

	type testCase struct {
		name       string
		getRoute   routeFactoryFn
		mwFunction []MiddlewareFunc
	}

	tt := []testCase{
		{
			name: "test for zero middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{},
		},
		{
			name: "test for one middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{func(ctx *Context, _ HandlerFunc) {
				ctx.writer.write([]byte("first-mw"))
			}},
		},
		{
			name: "test for two middleware",
			getRoute: func(t *testing.T) *Route {
				return newRoute("test-fun", testHandler)
			},
			mwFunction: []MiddlewareFunc{func(ctx *Context, _ HandlerFunc) {
				ctx.writer.write([]byte("first-mw"))
			},
				func(ctx *Context, _ HandlerFunc) {
					ctx.writer.write([]byte("second"))
				}},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			route := tc.getRoute(t)

			ctx := newContext(nil, nil)

			route.RegisterMiddlewares(tc.mwFunction...)

			var (
				expLen = len(tc.mwFunction) + 1
				gotLen = len(route.chain)
			)

			if gotLen != expLen {
				t.Errorf("expected chain length: %d; got length: %d\n", expLen, gotLen)
			}

			route.getHandler()(ctx)

			if !reflect.DeepEqual(ctx.writer.b, testW) {
				t.Errorf("expected writte value: %s; actually written value: %s\n", string(testW), string(ctx.writer.b))
			}
		})
	}
}

func TestGetPathParams(t *testing.T) {
	type testCase struct {
		name        string
		storedUrl   string
		incomingUrl string
		expected    []pathParam
	}

	tt := []testCase{
		{
			name:        "the functions returns empty slice if there is no path params",
			storedUrl:   "/foo/bar/baz",
			incomingUrl: "/foo/bar/baz",
			expected:    []pathParam{},
		},
		{
			name:        "the functions returns all the params with correct values",
			storedUrl:   "/foo/{bar}/{baz}",
			incomingUrl: "/foo/test-bar/test-baz",
			expected: []pathParam{
				{
					key:   "bar",
					value: "test-bar",
				},
				{
					key:   "baz",
					value: "test-baz",
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			params := getPathParams(tc.storedUrl, tc.incomingUrl)

			if !reflect.DeepEqual(params, tc.expected) {
				t.Errorf("missing values")
			}
		})
	}
}
