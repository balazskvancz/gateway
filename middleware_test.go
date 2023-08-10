package gateway

import (
	"net/http/httptest"
	"strings"
	"testing"
)

type mockHandler struct {
	hasBeenCalled bool
}

func (mh *mockHandler) handlerFunc(_ *Context) {
	mh.hasBeenCalled = true
}

type mockMatchingMiddleware struct {
	*matchingMiddleware
	calledSeq []string
}

func TestGetHandler(t *testing.T) {
	type (
		mockMatchingMiddlewareFactory func(*testing.T) *mockMatchingMiddleware
		mockHandlerFactory            func(*testing.T) *mockHandler
	)

	type testCase struct {
		name                  string
		getMatchingMiddleware mockMatchingMiddlewareFactory
		getHandler            mockHandlerFactory
		isHandlerCalled       bool
		expectedCalledSeq     []string
	}

	tt := []testCase{
		{
			name: "the handler is called if the matchingMiddleware is nil",
			getMatchingMiddleware: func(t *testing.T) *mockMatchingMiddleware {
				return &mockMatchingMiddleware{
					matchingMiddleware: nil,
					calledSeq:          make([]string, 0),
				}
			},
			getHandler: func(t *testing.T) *mockHandler {
				return &mockHandler{}
			},
			isHandlerCalled:   true,
			expectedCalledSeq: []string{},
		},
		{
			name: "the handler is not being called if the matchingMiddleware's pre mw breaks",
			getMatchingMiddleware: func(t *testing.T) *mockMatchingMiddleware {
				mockMw := &mockMatchingMiddleware{
					calledSeq: []string{},
				}

				m := newMiddleware(func(_ *Context, _ HandlerFunc) {
					mockMw.calledSeq = append(mockMw.calledSeq, "1")
				})

				mPost := newMiddleware(func(_ *Context, _ HandlerFunc) {
					mockMw.calledSeq = append(mockMw.calledSeq, "2")
				})

				mm := &matchingMiddleware{
					pre:  middlewareChain{m},
					post: middlewareChain{mPost},
				}

				mockMw.matchingMiddleware = mm

				return mockMw
			},
			getHandler: func(t *testing.T) *mockHandler {
				return &mockHandler{}
			},
			isHandlerCalled:   false,
			expectedCalledSeq: []string{"1", "2"},
		},
		{
			name: "the handler is not being called if some pre mw breaks",
			getMatchingMiddleware: func(t *testing.T) *mockMatchingMiddleware {
				mockMw := &mockMatchingMiddleware{
					calledSeq: []string{},
				}

				var (
					m1 = newMiddleware(func(ctx *Context, next HandlerFunc) {
						mockMw.calledSeq = append(mockMw.calledSeq, "1")
						next(ctx)
					})

					m2 = newMiddleware(func(_ *Context, _ HandlerFunc) {
						mockMw.calledSeq = append(mockMw.calledSeq, "2")
					})

					m3 = newMiddleware(func(_ *Context, _ HandlerFunc) {
						mockMw.calledSeq = append(mockMw.calledSeq, "3")
					})
				)

				mm := &matchingMiddleware{
					pre:  middlewareChain{m1, m2, m3},
					post: make(middlewareChain, 0),
				}

				mockMw.matchingMiddleware = mm

				return mockMw
			},
			getHandler: func(t *testing.T) *mockHandler {
				return &mockHandler{}
			},
			isHandlerCalled:   false,
			expectedCalledSeq: []string{"1", "2"},
		},
		{
			name: "the handler is being called if the matchingMiddleware's pre calls next",
			getMatchingMiddleware: func(t *testing.T) *mockMatchingMiddleware {
				mockMw := &mockMatchingMiddleware{
					calledSeq: []string{},
				}

				var (
					m1 = newMiddleware(func(ctx *Context, next HandlerFunc) {
						mockMw.calledSeq = append(mockMw.calledSeq, "1")
						next(ctx)
					})

					m2 = newMiddleware(func(ctx *Context, next HandlerFunc) {
						mockMw.calledSeq = append(mockMw.calledSeq, "2")
						next(ctx)
					})
				)

				mm := &matchingMiddleware{
					pre:  middlewareChain{m1, m2},
					post: make(middlewareChain, 0),
				}

				mockMw.matchingMiddleware = mm

				return mockMw
			},
			getHandler: func(t *testing.T) *mockHandler {
				return &mockHandler{}
			},
			isHandlerCalled:   true,
			expectedCalledSeq: []string{"1", "2"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				mm      = tc.getMatchingMiddleware(t)
				handler = tc.getHandler(t)

				recorder = httptest.NewRecorder()
			)

			handlerFunc := mm.matchingMiddleware.getHandler(handler.handlerFunc)

			ctx := newContext(nil, nil)

			ctx.reset(recorder, nil)

			handlerFunc(ctx)

			if tc.isHandlerCalled != handler.hasBeenCalled {
				t.Errorf("expected called to be: %t; got: %t\n", tc.isHandlerCalled, handler.hasBeenCalled)
			}

			var (
				expSeq = strings.Join(tc.expectedCalledSeq, ";")
				gotSeq = strings.Join(mm.calledSeq, ";")
			)

			if expSeq != gotSeq {
				t.Errorf("expected mw sequantial: %s; got: %s\n", expSeq, gotSeq)
			}
		})
	}

}
