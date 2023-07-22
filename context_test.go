package gateway

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestBasicRequest(t *testing.T) {
	tt := []struct {
		name string
		req  *http.Request

		expectedMethod                string
		expectedUrl                   string
		expectedUrlParts              []string
		expectedUrlWithOutQueryParams string
	}{
		{
			name: "http get method, without any query params",
			req:  httptest.NewRequest(http.MethodGet, "/api/foo/bar", nil),

			expectedMethod:                http.MethodGet,
			expectedUrl:                   "/api/foo/bar",
			expectedUrlParts:              []string{"api", "foo", "bar"},
			expectedUrlWithOutQueryParams: "/api/foo/bar",
		},

		{
			name: "http get method, with query params",
			req:  httptest.NewRequest(http.MethodGet, "/api/foo/bar?paramOne=1&paramTwo=2", nil),

			expectedMethod:                http.MethodGet,
			expectedUrl:                   "/api/foo/bar?paramOne=1&paramTwo=2",
			expectedUrlParts:              []string{"api", "foo", "bar"},
			expectedUrlWithOutQueryParams: "/api/foo/bar",
		},
		{
			name: "http post method, without any query params",
			req:  httptest.NewRequest(http.MethodPost, "/api/foo/bar", nil),

			expectedMethod:                http.MethodPost,
			expectedUrl:                   "/api/foo/bar",
			expectedUrlParts:              []string{"api", "foo", "bar"},
			expectedUrlWithOutQueryParams: "/api/foo/bar",
		},

		{
			name: "http post method, with query params",
			req:  httptest.NewRequest(http.MethodPost, "/api/foo/bar?paramOne=1&paramTwo=2", nil),

			expectedMethod:                http.MethodPost,
			expectedUrl:                   "/api/foo/bar?paramOne=1&paramTwo=2",
			expectedUrlParts:              []string{"api", "foo", "bar"},
			expectedUrlWithOutQueryParams: "/api/foo/bar",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newContext(getContextIdChannel())
			ctx.reset(nil, tc.req)

			if ctx.GetRequestMethod() != tc.expectedMethod {
				t.Errorf("expected method: %s; got: %s\n", tc.expectedMethod, ctx.GetRequestMethod())
			}

			if ctx.GetFullUrl() != tc.expectedUrl {
				t.Errorf("expected url: %s; got: %s\n", tc.expectedUrl, ctx.GetFullUrl())
			}

			if ctx.GetUrlWithoutQueryParams() != tc.expectedUrlWithOutQueryParams {
				t.Errorf("expected query less url: %s; got: %s\n", tc.expectedUrlWithOutQueryParams, ctx.GetUrlWithoutQueryParams())
			}

			gotUrlParts := ctx.GetUrlParts()
			if len(tc.expectedUrlParts) != len(gotUrlParts) {
				t.Fatalf("expected url parts length: %d; got: %d\n", len(tc.expectedUrlParts), len(gotUrlParts))
			}

			notOk := []string{}

			for idx, p := range gotUrlParts {
				if p != tc.expectedUrlParts[idx] {
					notOk = append(notOk, p)
				}
			}

			if len(notOk) != 0 {
				t.Errorf("expected every part to be equal")
			}

		})
	}
}

func TestSendJson(t *testing.T) {
	tt := []struct {
		name         string
		data         interface{}
		expectedBody []byte
	}{
		{
			name: "the right data was written",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "test",
				Age:  10,
			},
			expectedBody: []byte(`{"name":"test","age":10}`),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newContext(getContextIdChannel())
			ctx.reset(nil, nil)

			ctx.SendJson(tc.data)

			var (
				writtenCode = ctx.writer.statusCode
				writtenBody = ctx.writer.b
				// writtenHeader = ctx.writer.header
			)

			if !reflect.DeepEqual(tc.expectedBody, writtenBody) {
				t.Errorf("expected body: %s; got body: %s\n", string(tc.expectedBody), string(writtenBody))
			}

			if writtenCode != http.StatusOK {
				t.Errorf("expected statusCode: %d; got statusCode: %d\n", http.StatusOK, writtenCode)
			}
		})
	}
}

func TestSendNotFound(t *testing.T) {
	tt := []struct {
		name string
	}{
		{
			name: "the functions sends not found",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newContext(getContextIdChannel())
			ctx.reset(nil, nil)

			ctx.SendNotFound()

			writtenCode := ctx.writer.statusCode

			if writtenCode != http.StatusNotFound {
				t.Errorf("expected code: %d; got: %d\n", http.StatusNotFound, writtenCode)
			}
		})
	}
}

func TestSendOk(t *testing.T) {
	tt := []struct {
		name string
	}{
		{
			name: "the functions sends ok",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newContext(getContextIdChannel())
			ctx.reset(nil, nil)

			ctx.SendOk()

			writtenCode := ctx.writer.statusCode

			if writtenCode != http.StatusOK {
				t.Errorf("expected code: %d; got: %d\n", http.StatusOK, writtenCode)
			}
		})
	}
}
