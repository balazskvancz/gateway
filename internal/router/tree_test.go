package router

import (
	"fmt"
	"testing"

	"github.com/balazskvancz/gateway/internal/gcontext"
)

var emptyChain = createNewMWChain(func(g *gcontext.GContext) {})

func TestCreateNodeList(t *testing.T) {

	tt := []struct {
		name string

		url   string
		chain *middlewareChain

		expectedNode  *node
		expectedError error
	}{
		{
			name:  "the functions returns error, if mwChain is nil",
			chain: nil,
			url:   "",

			expectedNode:  nil,
			expectedError: errFnIsNil,
		},
		{
			name:  "the functions returns error, if url is empty",
			chain: emptyChain,
			url:   "",

			expectedNode:  nil,
			expectedError: errRouteToShort,
		},
		{
			name:  "the function returns the node, without any error",
			chain: emptyChain,
			url:   "/foo/bar",

			expectedNode:  &node{},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotErr := createNodeList(tc.url, tc.chain)

			if tc.expectedError != nil {
				if tc.expectedError != gotErr {
					t.Errorf("expected error: %v, got error: %v\n", tc.expectedError, gotErr)
				}
			} else {
				if gotErr != nil {
					t.Errorf("didnt expect error; got: %v\n", gotErr)
				}
			}

			if tc.expectedNode != nil {
				if gotNode == nil {
					t.Errorf("expected not nil node, but got one")
				}
			} else {
				if gotNode != nil {
					t.Errorf("expected nil node, but got one")
				}
			}
		})

	}
}

func TestAddToTree(t *testing.T) {
	tree := createTree()

	already, _ := createNodeList("/api/already/added", emptyChain)

	tree.addToTree(already)

	notStartWithApiNode, _ := createNodeList("/foo/bar", emptyChain)
	startsWithApiNode, _ := createNodeList("/api/foo/bar", emptyChain)

	tt := []struct {
		name          string
		node          *node
		expectedError error
	}{
		{
			name:          "the function returns error, if the given node doesnt start with /api",
			node:          notStartWithApiNode,
			expectedError: errMustStartWithApi,
		},
		{
			name:          "the function returns error, the given url is already inserted",
			node:          already,
			expectedError: errRouteAlreadyExits,
		},
		{
			name:          "the function returns nil error, if it could insert the node",
			node:          startsWithApiNode,
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotError := tree.addToTree(tc.node)

			if gotError != tc.expectedError {
				t.Errorf("expected error: %v; got error: %v\n", tc.expectedError, gotError)
			}

		})

	}

}

func TestFindNode(t *testing.T) {
	normalTree := createTree()
	emptyChildrenTree := createTree()

	route1, err1 := createNodeList("/api/foo/baz/bar", emptyChain)

	if err1 != nil {
		t.Errorf("route1 create error: %v\n", err1)
	}

	route2, err2 := createNodeList("/api/:id/bar/:id2", emptyChain)

	if err2 != nil {
		t.Errorf("route2 create error: %v\n", err2)
	}

	if err := normalTree.addToTree(route1); err != nil {
		t.Errorf("route1 add; got error: %v\n", err)
	}

	if err := normalTree.addToTree(route2); err != nil {
		t.Errorf("route2 add; got error: %v\n", err)
	}

	expectedParams := make(map[string]string)
	expectedParams["id"] = "1"
	expectedParams["id2"] = "2"

	tt := []struct {
		name string
		tree *tree

		url string

		expectedNode   *node
		expectedParams map[string]string
	}{
		{
			name:           "the function returns nil, nil if the tree is nil",
			tree:           nil,
			url:            "/foo",
			expectedNode:   nil,
			expectedParams: nil,
		},
		{
			name:           "the function returns nil, nil if the tree hasnt got any children",
			tree:           emptyChildrenTree,
			url:            "/foo",
			expectedNode:   nil,
			expectedParams: nil,
		},
		{
			name:           "the function returns nil, nil if the there is no matching route",
			tree:           normalTree,
			url:            "/foo",
			expectedNode:   nil,
			expectedParams: nil,
		},
		{
			name:           "the function returns node1, nil if look for route1s route",
			tree:           normalTree,
			url:            "/api/foo/baz/bar",
			expectedNode:   route1,
			expectedParams: nil,
		},
		{
			name:           "the function returns node2 and the good params if looking for route2 ",
			tree:           normalTree,
			url:            "/api/1/bar/2",
			expectedNode:   route2,
			expectedParams: expectedParams,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotParams := tc.tree.findNode(tc.url)

			if tc.expectedNode == nil {
				if gotNode != nil {
					t.Errorf("didnt expect node, but got one")
				}
			} else {
				if gotNode == nil {
					t.Errorf("didnt expect nil node, but got one")
				}
			}

			if len(tc.expectedParams) == 0 {
				if len(gotParams) > 0 {
					for k, r := range gotParams {
						fmt.Printf("%s => %s\n", k, r)
					}
					t.Errorf("expected no params, but got")
				}
			}
		})
	}

}
