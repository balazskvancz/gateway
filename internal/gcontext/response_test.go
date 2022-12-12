package gcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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

			ctx := New(recorder, nil)

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

	httptest.NewRecorder()
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

			ctx := New(rec, nil)

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

			ctx := New(rec, nil)

			ctx.SendOk()

			if rec.Code != http.StatusOK {
				t.Errorf("expected code: %d; got: %d\n", http.StatusOK, rec.Code)
			}
		})
	}
}
