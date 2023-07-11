package gateway

import (
	"errors"
	"strings"
	"sync"
)

//
// URLs are stored in the form of: /api/foo/{resource}/{id}
// where „resource” and „id” are the two path params
// of the request.
//
// Examples of matching urls for the scheme above:
//	/api/foo/products/1										-> resource="products" 		id="1"
// 	/api/foo/categories/example-category 	-> resource="categories"  id="example-category"
//
// Due to the way we store these routes, there is a chance of trying to store
// ambigoues routes aswell. That obviously would cause a bad behaviour.
// Example for these routes:
//
// /api/{resource}/get
// /api/products/get
//
// The two above would cause an error, however the two 2 below would not:
//
// /api/{resource}/get
// /api/products/get-all
//

const (
	slash = '/'

	curlyStart = '{'
	curlyEnd   = '}'
)

type (
	predicateFunction func(*node) bool
)

type tree struct {
	mu   sync.RWMutex
	root *node
}

type node struct {
	key   string
	value any

	children []*node
}

func (n *node) isLeaf() bool {
	return n.value != nil
}

func newTree() *tree {
	return &tree{
		mu: sync.RWMutex{},
	}
}

// insert tries to store a key-value pair in the tree.
// In case of unsuccessful insertion, we return the root of the error.
func (t *tree) insert(key string, value any) error {
	if t == nil {
		return errTreeIsNil
	}

	if key == "" {
		return errKeyIsEmpty
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if err := checkUrl(key); err != nil {
		return err
	}

	// If the root is still nil, then the new node is the root.
	if t.root == nil {
		t.root = createNewNode(key, value)
		return nil
	}

	return insertRec(t.root, key, value)
}

// iterateInsert iterates on the given node's children, and calls
// insertRec on each one. If there is no error during the recursive calls
// we successfully inserted the new node. Otherwise, if get an error that
// differs from errNoCommonPrefix, we return it. If none of those happaned, we
// simply return errNoCommonPrefix which indicates we were trying to
// insert on a wrong branch.
func iterateInsert(n *node, key string, value any) error {
	for _, ch := range n.children {
		insertErr := insertRec(ch, key, value)

		if insertErr == nil {
			return nil
		}

		if !errors.Is(insertErr, errNoCommonPrefix) {
			return insertErr
		}
	}

	return errNoCommonPrefix
}

// insertRec
func insertRec(n *node, key string, value any) error {
	lcp := longestCommonPrefix(n.key, key)

	// There is no chance of inserting in this branch.
	if lcp == 0 {
		return errNoCommonPrefix
	}

	keyLen := len(key)

	// If the length of the common part is equal to the inserting key,
	// then the current node is place we wanted to insert in the first place.
	if lcp == keyLen {
		// If it is already leaf, return error.
		if n.isLeaf() {
			return errKeyIsAlreadyStored
		}

		// Otherwise we simply the store the value and we are done.
		n.value = value

		return nil
	}

	// Three other possibilities:
	// 		1) the current node's key is longer than the LCP => must split keys,
	// 		2) current node's are same as lcp, and new key is longer =>,
	// 		3) otherwise the new node should be amongs the children of the current node.

	if len(n.key) > lcp {
		var (
			cNewNode = createNewNode(n.key[lcp:], n.value, n.children...)
			newNode  = createNewNode(key[lcp:], value)
		)

		n.value = nil
		n.key = n.key[:lcp]
		n.children = []*node{cNewNode, newNode}

		return nil
	}

	keyRem := key[lcp:]

	err := iterateInsert(n, keyRem, value)

	if err == nil {
		return nil
	}

	if !errors.Is(err, errNoCommonPrefix) {
		return err
	}

	addToChildren(n, createNewNode(keyRem, value))

	return nil
}

func addToChildren(n, newNode *node) {
	n.children = append(n.children, newNode)
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

	// Check for path params, and check for its syntax.
	return checkPathParams(url)
}

func checkPathParams(url string) error {
	// If there is none of the curly brackets, we are good to go.
	if !strings.ContainsRune(url, curlyStart) && !strings.ContainsRune(url, curlyEnd) {
		return nil
	}

	var (
		insideParam = false
		counter     = 0
	)

	for counter < len(url) {
		if url[counter] == slash {
			// If we are inside a path param, there cant be a slash.
			if insideParam {
				return errBadPathParamSyntax
			}
		}

		if url[counter] == curlyStart {
			if insideParam {
				return errBadPathParamSyntax
			}

			insideParam = true
		}

		if url[counter] == curlyEnd {
			if !insideParam {
				return errBadPathParamSyntax
			}
			insideParam = false
		}

		counter++
	}

	// If we are still inside a path param
	// after the url is ended, means error.
	if insideParam {
		return errBadPathParamSyntax
	}

	return nil
}

// checkTree does a basic check on the given tree, returns error
// if either the tree or the root is nil.
func checkTree(t *tree) error {
	if t == nil {
		return errTreeIsNil
	}

	if t.root == nil {
		return errRootIsNil
	}

	return nil
}

// min returns the minimum of two given numbers.
func min(num1, num2 int) int {
	if num1 > num2 {
		return num2
	}

	return num1
}

// longestCommonPrefix returns the length of the
// longest common prefix of two given strings.
func longestCommonPrefix(str1, str2 string) int {
	var counter = 0

	maxVal := min(len(str1), len(str2))

	for counter < maxVal && str1[counter] == str2[counter] {
		counter += 1
	}

	return counter
}

// createNewNode is a factory for creating new nodes.
func createNewNode(key string, value any, children ...*node) *node {
	n := &node{
		key:      key,
		value:    value,
		children: make([]*node, 0),
	}

	if len(children) > 0 {
		n.children = children
	}

	return n
}

// find starts the search for given key and returns a pointer to
// the found node. If there is no match, it returns nil.
func (t *tree) find(key string) *node {
	if err := checkTree(t); err != nil {
		return nil
	}

	if key == "" {
		return nil
	}

	return findRec(t.root, key, false)
}

// findRec is the main logic for conducting the search in a recursive manner.
// It looks for match on the given node's level, and calls itself recursively
// amongs its children, until the search is over.
func findRec(n *node, key string, isWildcard bool) *node {
	if n == nil {
		return nil
	}

	// If the current node's key contains curlyStart char,
	// that means there is a start of wildcard part.
	if strings.ContainsRune(n.key, curlyStart) {
		isWildcard = true
	}

	lcp := longestCommonPrefix(n.key, key)

	// If there is nothing in common and it is not wildcard, then we are off.
	if lcp == 0 && !isWildcard {
		return nil
	}

	// In case of non wildcard part, normal string comp.
	if !isWildcard {
		if key == n.key {
			return n
		}

		// If the current node's key is longer than the lcp, no match.
		if lcp < len(n.key) {
			return nil
		}

		// Otherwise have to look amongst the children recursively.
		for _, c := range n.children {
			if found := findRec(c, key[lcp:], isWildcard); found != nil {
				return found
			}
		}

		return nil
	}

	var (
		nodeKeyRem   = n.key[lcp:]
		searchKeyRem = key[lcp:]
	)

	offset1, offset2, isStillWildcard := getOffsets(nodeKeyRem, searchKeyRem, true)

	// Meaning we didnt shift until the last char, not a full match in this level.
	if len(nodeKeyRem) != offset1 {
		return nil
	}

	newSearchKey := searchKeyRem[offset2:]

	// If there is nothing from the original search key
	// we are on the exact node we were looking for.
	if newSearchKey == "" {
		// Only to check if this node is a leaf, or not.
		if n.isLeaf() {
			return n
		}
		return nil
	}

	// Have to continue search on the next level.
	for _, ch := range n.children {
		if found := findRec(ch, newSearchKey, isStillWildcard); found != nil {
			return found
		}
	}

	return nil
}

// getOffsets returns the offset of the first and second given string and whether it is still
// a wildcard search. These offsets are displaying how far should each string be shifted, how long
// is the common part including wildcard option.
func getOffsets(storedKey, searchKey string, isWildcard bool) (int, int, bool) {
	var (
		i = 0
		j = 0

		storedKeyLen = len(storedKey)
		searchKeyLen = len(searchKey)
	)

	for {
		if i >= storedKeyLen {
			break
		}

		if j >= searchKeyLen && !isWildcard {
			break
		}

		if storedKey[i] == curlyStart {
			isWildcard = true
			i++
			continue
		}

		// In case of closing a {id} part, we have to
		// move forward in the search key aswell.
		if storedKey[i] == curlyEnd {
			isWildcard = false

			cSearchRem := searchKey[j:]

			nextSlashIdx := strings.IndexRune(cSearchRem, slash)

			j += func() int {
				// There is no other / remaining.
				if nextSlashIdx == -1 {
					return len(cSearchRem)
				}
				// Otherwise skip that amount.
				return nextSlashIdx
			}()

			i++

			continue
		}

		// If we are inside of a wildcard check,
		// we only increment the stored keys counter.
		if isWildcard {
			i++
			continue
		}

		if storedKey[i] != searchKey[j] {
			break
		}

		i++
		j++
	}

	return i, j, isWildcard
}

// findLongestMatch is similar to find but it doesnt include any wildcard params at all.
// And it is not looking for perfect match, rather it finds the longest „route” based on the given string.
// Used for storing services based on their prefixes.
func (t *tree) findLongestMatch(key string) *node {
	if err := checkTree(t); err != nil {
		return nil
	}

	if key == "" {
		return nil
	}

	return findLongestMatchRec(t.root, key)
}

// findLongestMatchRec
func findLongestMatchRec(n *node, key string) *node {
	if n == nil {
		return nil
	}

	lcp := longestCommonPrefix(n.key, key)

	if lcp == 0 {
		return nil
	}

	if lcp != len(n.key) {
		return nil
	}

	for _, ch := range n.children {
		if node := findLongestMatchRec(ch, key[lcp:]); node != nil {
			return node
		}
	}

	if !n.isLeaf() {
		return nil
	}

	return n
}

// getAllLeaf returns all of leaf nodes.
func (t *tree) getAllLeaf() []*node {
	if err := checkTree(t); err != nil {
		return nil
	}

	return getAllLeafRec(t.root)
}

func getAllLeafRec(n *node) []*node {
	arr := make([]*node, 0)

	for _, c := range n.children {
		chArr := getAllLeafRec(c)

		if len(chArr) > 0 {
			arr = append(arr, chArr...)
		}
	}

	if n.isLeaf() {
		arr = append(arr, n)
	}

	return arr
}

// getByPredicate does a search in the tree based on given function.
// It uses DFS as the algorithm to traverse the tree.
func (t *tree) getByPredicate(fn predicateFunction) *node {
	if err := checkTree(t); err != nil {
		return nil
	}

	return getByPredicateRec(t.root, fn)
}

func getByPredicateRec(n *node, fn predicateFunction) *node {
	if n == nil {
		return nil
	}

	if fn(n) {
		return n
	}

	for _, ch := range n.children {
		if match := getByPredicateRec(ch, fn); match != nil {
			return match
		}
	}

	return nil
}
