package gateway

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	_RouteLength = 5
)

const noFoudRoute = "/api/bar/foo/dsa"

var _PossibleUrlParts = []string{
	"foo",
	"bar",
	"baz",
	"lorem",
	"ipsum",
	"testing",
	"the",
	"gateway",
	"thesis",
	":id",
}

func BenchmarkTree(b *testing.B) {

	tt := []struct {
		n int
	}{
		{
			n: 10,
		},
		{
			n: 50,
		},
		{
			n: 100,
		},
		{
			n: 300,
		},
		{
			n: 500,
		},
	}

	for _, tc := range tt {
		tree := createTree()

		routes := test_createRoutes(tc.n, &[]string{})

		for _, rout := range *routes {
			n, err := createNodeList(rout, emptyChain)

			if err != nil {
				b.Fatalf("expected no error; got: %v\n", err)
			}

			if err := tree.addToTree(n); err != nil {
				b.Fatalf("expected no error; got: %v\n", err)
			}
		}

		name := fmt.Sprintf("testing with %d routes", tc.n)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				// index := rand.Intn(len(*routes))
				// tree.findNode(test_normalizeRoute((*routes)[index]))
				tree.findNode(noFoudRoute)
			}
		})
	}

}

// Test helper, that creates the given amount of routes.
func test_createRoutes(n int, arr *[]string) *[]string {
	if len(*arr) == n {
		return arr
	}

	route := test_createRoute()

	if test_contains(route, arr) {
		return test_createRoutes(n, arr)
	}

	*arr = append(*arr, route)

	return test_createRoutes(n, arr)
}

func test_createRoute() string {
	route := "/api"
	for i := 0; i < _RouteLength; i++ {

		index := rand.Intn(len(_PossibleUrlParts))

		route += "/" + _PossibleUrlParts[index]
	}

	return route
}

func test_contains(str string, arr *[]string) bool {
	for _, e := range *arr {
		if e == str {
			return true
		}
	}

	return false
}

// func test_normalizeRoute(str string) string {
// if !strings.Contains(str, ":id") {
// return str
// }

// return strings.ReplaceAll(str, ":id", "1")
// }

func BenchmarkPool(b *testing.B) {
	router := newRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		router._poolTester(writer, req)
	}
}

func BenchmarkContext(b *testing.B) {
	router := newRouter(nil)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	writer := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		router._createTester(writer, req)
	}
}
