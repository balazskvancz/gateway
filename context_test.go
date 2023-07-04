package gateway

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
		expectedBody string
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
			expectedBody: `{"name":"test","age":10}`,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			ctx := newContext(getContextIdChannel())
			ctx.reset(recorder, nil)

			ctx.SendJson(tc.data)

			if recorder.Code != http.StatusOK {
				t.Errorf("expected http code: %d; got code: %d\n", http.StatusOK, recorder.Code)
			}

			if recorder.Body.String() != tc.expectedBody {
				t.Errorf("expected body: %s; got body: %s\n", tc.expectedBody, recorder.Body.String())
			}

			gotContentType := recorder.HeaderMap.Get("Content-Type")

			if gotContentType != JsonContentType {
				t.Errorf("expected content-type: %s; got: %s\n", JsonContentType, gotContentType)
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
			rec := httptest.NewRecorder()

			ctx := newContext(getContextIdChannel())
			ctx.reset(rec, nil)

			ctx.SendNotFound()

			if rec.Code != http.StatusNotFound {
				t.Errorf("expected code: %d; got: %d\n", http.StatusNotFound, rec.Code)
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
			rec := httptest.NewRecorder()

			ctx := newContext(getContextIdChannel())
			ctx.reset(rec, nil)

			ctx.SendOk()

			if rec.Code != http.StatusOK {
				t.Errorf("expected code: %d; got: %d\n", http.StatusOK, rec.Code)
			}
		})
	}
}
