package gcontext

import (
	"net/http"
	"net/http/httptest"
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
			ctx := New(nil, tc.req)

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
