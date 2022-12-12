package communicator

import (
	"net/http"
	"testing"
	"time"
)

var urls = []string{
	"/api/mock-1",
	"/api/mock-3",
}

type Client1S struct{}

func (c Client1S) GoRequest(method, url string, data []byte, chans *concurrentChan) {
	time.Sleep(1 * time.Second)

	// The same url, that we give.
	chans.url <- url

	chans.GChan.Status <- http.StatusOK
	chans.GChan.Header <- http.Header{}
	chans.GChan.Data <- []byte(`{"ok": true}`)
}

type Client3S struct{}

func (c Client3S) GoRequest(method, url string, data []byte, chans *concurrentChan) {
	time.Sleep(3 * time.Second)

	// The same url, that we give.
	chans.url <- url

	chans.GChan.Status <- http.StatusBadRequest
	chans.GChan.Header <- http.Header{}
	chans.GChan.Data <- []byte(`{"ok": false}`)
}

func TestConcurrentAll(t *testing.T) {
	const mockMethod = http.MethodGet
	var mockBody []byte = nil

	expectedMap := make(map[string]GChanValues)

	expectedMap[urls[0]] = GChanValues{
		Data:       []byte(`{"ok": true}`),
		StatusCode: http.StatusOK,
		Header:     http.Header{},
	}

	expectedMap[urls[1]] = GChanValues{
		Data:       []byte(`{"ok": false}`),
		StatusCode: http.StatusBadRequest,
		Header:     http.Header{},
	}

	tt := []struct {
		name     string
		reqs     []ConcurrentRequest
		expected ConcurrentResponse
	}{
		{
			name: "the functions awaits all the responses and returns the expected values",
			reqs: []ConcurrentRequest{
				{
					Client: Client1S{},
					Url:    urls[0],
					Method: mockMethod,
					Data:   mockBody,
				},
				{
					Client: Client3S{},
					Url:    urls[1],
					Method: mockMethod,
					Data:   mockBody,
				},
			},
			expected: expectedMap,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cResponse := ConcurrentAll(tc.reqs...)

			if len(cResponse) != len(expectedMap) {
				t.Errorf("expected length of response: %d; got length: %d\n", len(expectedMap), len(cResponse))
			}

			for _, url := range urls {
				gotData, ok := cResponse[url]

				if !ok {
					t.Fatalf("expected to have: %s; got not nil\n", url)
				}

				expectedData, _ := expectedMap[url]

				gotStr := string(gotData.Data)
				expectedStr := string(expectedData.Data)

				if gotStr != expectedStr {
					t.Errorf("expected body: %s; got body: %s\n", expectedStr, gotStr)
				}

				if gotData.StatusCode != expectedData.StatusCode {
					t.Errorf("expected statusCode: %d; got statusCode: %d\n", expectedData.StatusCode, gotData.StatusCode)
				}
			}
		})
	}
}
