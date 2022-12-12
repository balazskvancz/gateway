package router

import (
	"errors"
	"fmt"
	"strings"

	"github.com/balazskvancz/gateway/pkg/utils"
)

var (
	errConfusingRoute    = errors.New("rotue is confusing")
	errFnIsNil           = errors.New("handlerfunc is nil")
	errRootIsNil         = errors.New("root is nil")
	errRouteAlreadyExits = errors.New("route alreay exists")
	errRouteToShort      = errors.New("route must be 2 path long")
)

type tree struct {
	root *node
}

type node struct {
	part      string
	childrens []*node
	mwChain   *middlewareChain
}

// Creates a new tree.
func createTree() *tree {
	return &tree{
		root: createRootNode(),
	}
}

// Creates the root, of empty tree.
func createRootNode() *node {
	return &node{
		part:      rootPrefix, // It always contains the root prefix.
		childrens: []*node{},  // Initializing an emtpy slice of nodes.
	}
}

// Creates a list of nodes by the given url.
// The last element will hold its HandlerFunc.
func createNodeList(url string, mwChain *middlewareChain) (*node, error) {
	if mwChain == nil {
		return nil, errFnIsNil
	}

	urlParts := utils.GetUrlParts(url)

	if len(urlParts) < 2 {
		return nil, errRouteToShort
	}

	// Should not include "/api".
	root := createNodeRecursively(urlParts, mwChain)

	if root == nil {
		return nil, errRootIsNil
	}

	return root, nil
}

// Helper function that creates the nodes recursively.
func createNodeRecursively(urlParts []string, mwChain *middlewareChain) *node {
	n := &node{
		part: urlParts[0],
	}

	// If there is more parts, we create its children recursively.
	if len(urlParts) > 1 {
		n.childrens = []*node{
			createNodeRecursively(urlParts[1:], mwChain),
		}
	}

	// Meaning we are the bottom of the tree.
	if len(urlParts) == 1 {
		n.mwChain = mwChain
	}

	return n
}

// Adds a node to an already existing tree.
func (t *tree) addToTree(n *node) error {
	root := t.root

	// Rules:
	// The trees root value must be the same as the given nodes value.
	// If there is a match, should return duplicateRouteErr

	if root.part != n.part {
		return errMustStartWithApi
	}

	if err := addToNode(root, n.childrens[0]); err != nil {
		return err
	}

	return nil
}

// Adds a node, to a leaf.
func addToNode(treeLeaf *node, n *node) error {
	//
	var matchingEl *node

	for _, ch := range treeLeaf.childrens {
		if ch.part == n.part {
			matchingEl = ch
		}
	}

	// We we didnt find any matching part at this level,
	// we should create one, and add it to the leaf.
	if matchingEl == nil {
		treeLeaf.childrens = append(treeLeaf.childrens, n)

		return nil
	}

	// If there is match, and this is the last, it means, we alreay have this route.
	if n.mwChain != nil {
		return errRouteAlreadyExits
	}

	// If there the node hasnt got any more children,
	// but the tobeadded node is not the last one,
	// should return error, because the path is "confusing."
	//
	// eg1 => /api/foo			-> this route already registered
	// eg2 => /api/foo/bar	-> this we want to register
	if len(matchingEl.childrens) == 0 && n.mwChain == nil {

	}

	// Now, we should continue, by calling this fn again
	// with the found node, the insertable nodes child.
	return addToNode(matchingEl, n.childrens[0])
}

// ----------------
// |     Walk     |
// ----------------

// Returns the the node and a [:key] => :value map with the route params.
func (t *tree) findNode(url string) (*node, map[string]string) {
	if t == nil {
		return nil, nil
	}

	if len(t.root.childrens) == 0 {
		return nil, nil
	}

	uParts := utils.GetUrlParts(url)

	lastNode, params := walkTree(t.root, uParts)

	if lastNode == nil {
		return nil, nil
	}

	return lastNode, params
}

func walkTree(n *node, parts []string) (*node, map[string]string) {
	if len(parts) == 0 {
		return nil, nil
	}

	params := make(map[string]string)
	currPart := parts[0]

	isRouteParam := strings.HasPrefix(n.part, ":")

	// If the current nodes part doesnt match and it isnt a route param, return nil.
	if n.part != currPart && !isRouteParam {
		return nil, nil
	}

	if isRouteParam {
		key := n.part[1:]

		params[key] = currPart
	}

	var el *node

	// If the current node, has a HandlerFunc,
	// and we are at the routeParts last element,
	// it means, we found a match.
	if n.mwChain != nil && len(parts) == 1 {
		return n, params
	}

	for _, ch := range n.childrens {
		foundNode, p := walkTree(ch, parts[1:])

		if foundNode != nil {
			el = foundNode
			params = p
		}
	}

	return el, params
}

func (t *tree) getRoutes() {
	getRoutes(t.root)
}

func getRoutes(n *node) {
	fmt.Println(n.part)

	for _, c := range n.childrens {
		getRoutes(c)
	}
}
