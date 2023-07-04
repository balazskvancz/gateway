package gateway

import (
	"errors"
	"testing"
)

type getTreeFn func(t *testing.T) *tree

func getHandler() any {
	type h struct{}

	return &h{}
}

func TestGetOffSets(t *testing.T) {
	tt := []struct {
		name       string
		storedKey  string
		searchKey  string
		isWildcard bool

		expectedOffset1    int
		expectedOffset2    int
		expectedIsWildcard bool
	}{
		{
			name:               "no wildcard and no match",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    0,
			expectedOffset2:    0,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, but wildcard by def (and still wildcard)",
			storedKey:          "foo",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    3,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},
		{
			name:               "no strict match, but wildcard by def (and not wildcard anymore)",
			storedKey:          "foo}",
			searchKey:          "bar",
			isWildcard:         true,
			expectedOffset1:    4,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and not wildcard anymore)",
			storedKey:          "{foo}",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    5,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
		{
			name:               "no strict match, wildcard but not by def (and still wildcard)",
			storedKey:          "{foo",
			searchKey:          "bar",
			isWildcard:         false,
			expectedOffset1:    4,
			expectedOffset2:    0,
			expectedIsWildcard: true,
		},

		{
			name:               "strict match and wildcard but not by def (and not wildcard anymore)",
			storedKey:          "/foo/{id}",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},

		{
			name:               "strict match and wildcard but not by def (and still wildcard)",
			storedKey:          "/foo/{id",
			searchKey:          "/foo/5/6",
			isWildcard:         false,
			expectedOffset1:    8,
			expectedOffset2:    5,
			expectedIsWildcard: true,
		},
		{
			name:               "strict match then wildcard then strict match again",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/bar",
			isWildcard:         false,
			expectedOffset1:    9,
			expectedOffset2:    6,
			expectedIsWildcard: false,
		},
		{
			name:               "strict match then wildcard then not strict match",
			storedKey:          "/{id}/bar",
			searchKey:          "/5/foo",
			isWildcard:         false,
			expectedOffset1:    6,
			expectedOffset2:    3,
			expectedIsWildcard: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotOffset1, gotOffset2, gotIsWildcard := getOffsets(tc.storedKey, tc.searchKey, tc.isWildcard)

			if gotOffset1 != tc.expectedOffset1 {
				t.Errorf("expected offset1: %d; got: %d\n", tc.expectedOffset1, gotOffset1)
			}

			if gotOffset2 != tc.expectedOffset2 {
				t.Errorf("expected offset2: %d; got: %d\n", tc.expectedOffset2, gotOffset2)
			}

			if gotIsWildcard != tc.expectedIsWildcard {
				t.Errorf("expected isWildcard: %v; got: %v\n", tc.expectedIsWildcard, gotIsWildcard)
			}
		})
	}
}

func TestTreeInsert(t *testing.T) {
	tt := []struct {
		name    string
		getTree getTreeFn
		input   string
		err     error
	}{
		{
			name:    "error if the tree is <nil>",
			getTree: func(t *testing.T) *tree { return nil },
			input:   "",
			err:     errTreeIsNil,
		},
		{
			name: "error if given url (key) is empty",
			getTree: func(t *testing.T) *tree {
				return newTree()
			},
			input: "",
			err:   errKeyIsEmpty,
		},
		{
			name: "error if given url (key) is not starting with a slash",
			getTree: func(t *testing.T) *tree {
				return newTree()
			},
			input: "foo",
			err:   errMissingSlashPrefix,
		},
		{
			name: "error if given url (key) is ending with a slash",
			getTree: func(t *testing.T) *tree {
				return newTree()
			},
			input: "/foo/",
			err:   errPresentSlashSuffix,
		},
		{
			name: "no error if insertion was successfull (empty tree)",
			getTree: func(t *testing.T) *tree {
				return newTree()
			},
			input: "/foo",
			err:   nil,
		},
		{
			name: "no error if insertion was successfull (not empty tree)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar", getHandler()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.insert("/foo/baz", getHandler()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/foo",
			err:   nil,
		},
		{
			name: "error on duplicate keys",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar", getHandler()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.insert("/foo/baz", getHandler()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.insert("/foo", getHandler()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/foo",
			err:   errKeyIsAlreadyStored,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			err := tree.insert(tc.input, getHandler())

			if tc.err != nil && !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got: %v\n", tc.err, err)
			}

			if tc.err == nil && err != nil {
				t.Errorf("unexpected error: %v\n", err)
			}

		})
	}

}

func TestTreeFind(t *testing.T) {
	tt := []struct {
		name      string
		getTree   func(t *testing.T) *tree
		searchKey string
		isExists  bool
	}{
		{
			name:      "cant find, if tree is <nil>",
			getTree:   func(t *testing.T) *tree { return nil },
			searchKey: "/foo",
			isExists:  false,
		},
		{
			name: "cant find, if root of tree is <nil>",
			getTree: func(t *testing.T) *tree {
				return &tree{}
			},
			searchKey: "/foo",
			isExists:  false,
		},
		{
			name: "cant find, if the search key is empty",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/api/foo", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "",
			isExists:  false,
		},

		// Simply not wildcard test.
		{
			name: "normal search without any wildcard (no match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar/baz", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bak",
			isExists:  false,
		},
		{
			name: "normal search without any wildcard (no match, only common subpart)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar/baz", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar",
			isExists:  false,
		},
		{
			name: "normal search without any wildcard (match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar/baz", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bar",
			isExists:  true,
		},

		// search with wildcard param
		{
			name: "wildcard search (no match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/{id}/baz", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/bar/bak",
			isExists:  false,
		},
		{
			name: "wildcard search - param is at the start (match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/{id}/baz/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bak/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/for/bak/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/1/baz/bar",
			isExists:  true,
		},
		{
			name: "wildcard search - param is in the middle (match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/{id}/baz", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/1/baz",
			isExists:  true,
		},
		{
			name: "wildcard search - param is at the end (match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/baz/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bak/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/for/bak/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/foo/bar/bar", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/foo/baz/1",
			isExists:  true,
		},

		// Multiple params
		{
			name: "wildcard search - multiple params (no match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/api/{resource}/get/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/api/{resource}/delete/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/api/{resource}/update/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/api/products/insert/1",
			isExists:  false,
		},
		{
			name: "wildcard search - multiple params (match)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/api/{resource}/get/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/api/{resource}/delete/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				if err := tree.insert("/api/{resource}/update/{id}", getHandler()); err != nil {
					t.Fatalf("not expected error, but got: %v\n", err)
				}

				return tree
			},
			searchKey: "/api/products/delete/1",
			isExists:  true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			node := tree.find(tc.searchKey)

			if tc.isExists && node == nil {
				t.Errorf("expected to find, but got <nil>")
			}

			if !tc.isExists && node != nil {
				t.Errorf("expected not to find, but got route")
			}
		})
	}
}

func TestCheckPathParams(t *testing.T) {
	tt := []struct {
		name  string
		input string
		err   error
	}{
		{
			name:  "no error, if there is no path params at all",
			input: "/foo/bar/baz",
			err:   nil,
		},
		{
			name:  "error if there no closing of param",
			input: "/foo/bar/{baz",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if there no start of param",
			input: "/foo/bar/baz}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if multiple start of param",
			input: "/foo/bar/{{baz",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if multiple end of param",
			input: "/foo/bar/baz}}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "error if there is a slash inside a param",
			input: "/foo/bar{/baz}",
			err:   errBadPathParamSyntax,
		},
		{
			name:  "no error if one path param",
			input: "/foo/bar/{baz}",
			err:   nil,
		},
		{
			name:  "no error if multiple path param",
			input: "/{foo}/bar/{baz}",
			err:   nil,
		},
		{
			name:  "error if one good and one bad param",
			input: "/{foo}/bar/baz}",
			err:   errBadPathParamSyntax,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := checkPathParams(tc.input); !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got: %v\n", tc.err, err)
			}
		})
	}
}