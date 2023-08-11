package gateway

import (
	"fmt"
	"math/rand"
	"strings"
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
	"{id}",
}

var toBeFound = true

func BenchmarkTree(b *testing.B) {
	tt := []int{10, 50, 100, 300, 500}

	for _, tc := range tt {
		tree := newTree[*Route]()

		routes := testCreateRoutes(tc, []string{})

		for _, r := range routes {
			route := newRoute(r, nil)

			if err := tree.insert(r, route); err != nil {
				b.Fatalf("expected no error; got: %v\n", err)
			}
		}

		name := fmt.Sprintf("testing with %d routes", tc)

		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if toBeFound {
					var (
						index     = rand.Intn(len(routes))
						searchurl = routes[index]
					)

					if node := tree.find(testNormalizeRoute(searchurl)); node == nil {
						b.Fatal("not found node; supposed to")
					}
				}

				if node := tree.find(noFoudRoute); node != nil {
					b.Fatal("found node; not supposed to")
				}
			}
		})
	}

}

func testCreateRoutes(n int, arr []string) []string {
	if len(arr) == n {
		return arr
	}

	newRoute := testCreateRoute()

	if includes(arr, newRoute) {
		return testCreateRoutes(n, arr)
	}

	arr = append(arr, newRoute)

	return testCreateRoutes(n, arr)
}

func testCreateRoute() string {
	route := "/api"
	for i := 0; i < _RouteLength; i++ {
		index := rand.Intn(len(_PossibleUrlParts))
		route += "/" + _PossibleUrlParts[index]
	}
	return route
}

func testNormalizeRoute(str string) string {
	if !strings.Contains(str, "{id}") {
		return str
	}

	return strings.ReplaceAll(str, "{id}", "1")
}
