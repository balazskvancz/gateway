package gateway

import (
	"errors"
	"testing"
)

type getTreeFn func(t *testing.T) *tree

func getTestservice() *service {
	return newService(nil)
}

func TestTreeInsert(t *testing.T) {
	type testCase struct {
		name    string
		getTree getTreeFn
		input   string
		err     error
	}

	tt := []testCase{
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
			name: "no error if insertion was successful (not empty tree)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo/bar", getTestservice()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				if err := tree.insert("/foo/baz", getTestservice()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/foo",
			err:   nil,
		},
		{
			name: "no error if insertion successful similar keys (not empty tree)",
			getTree: func(t *testing.T) *tree {
				tree := newTree()

				if err := tree.insert("/foo", getTestservice()); err != nil {
					t.Fatalf("unexpected error: %v\n", err)
				}

				return tree
			},
			input: "/fo",
			err:   nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			tree := tc.getTree(t)

			err := tree.insert(tc.input, getTestservice())

			if tc.err != nil && !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got: %v\n", tc.err, err)
			}

			if tc.err == nil && err != nil {
				t.Errorf("unexpected error: %v\n", err)
			}

		})
	}

}
