package gateway

import (
	"github.com/balazskvancz/rtree"
)

const (
	slash byte = '/'
)

type tree struct {
	*rtree.Tree[*service]
}

func newTree() *tree {
	return &tree{
		Tree: rtree.New[*service](),
	}
}

// insert tries to store a key-value pair in the tree.
// In case of unsuccessful insertion, we return the root of the error.
func (t *tree) insert(key string, value *service) error {
	if t == nil {
		return errTreeIsNil
	}

	if key == "" {
		return errKeyIsEmpty
	}

	if err := checkUrl(key); err != nil {
		return err
	}

	return t.Tree.Insert(key, value)
}

// checkUrl checks the given of errors such as missing slash prefix
// or bad path params.
func checkUrl(url string) error {
	// Leading slash.
	if url[0] != slash {
		return errMissingSlashPrefix
	}
	// Trailing slash.
	if url[len(url)-1] == slash {
		return errPresentSlashSuffix
	}
	return nil
}
