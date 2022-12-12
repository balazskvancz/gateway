package router

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/balazskvancz/gateway/pkg/gcontext"
	"github.com/balazskvancz/gateway/pkg/mock"
	"github.com/balazskvancz/gateway/pkg/registry"
)

const (
	rootPrefix = "api" // Every single route must start with "/api".
)

var (
	errMustStartWithApi = errors.New(`route must start with "/api"`)
)

type HandlerFunc func(*gcontext.GContext)

type Router struct {
	methodTrees    map[string]*tree
	serviceRegisty *registry.Registry
	contextPool    sync.Pool

	mockTree *tree

	middlewares map[string]HandlerFunc

	notFoundHandler HandlerFunc

	optionsHandler HandlerFunc
}

var _ http.Handler = (*Router)(nil)

// Creates a new instance of the router and returns a pointer to it.
func New(registry *registry.Registry) *Router {
	if registry == nil {
		return nil
	}

	return &Router{
		methodTrees: make(map[string]*tree),
		contextPool: sync.Pool{
			New: func() interface{} { return new(gcontext.GContext) },
		},
		serviceRegisty: registry,
		middlewares:    make(map[string]HandlerFunc), // Empty slice.
	}
}

// Setting notFoundHandler for Router.
func (router *Router) SetNotFoundHandler(handler HandlerFunc) {
	router.notFoundHandler = handler
}

// Setting optionsFoundHandler for Router.
func (router *Router) SetOptionsHandler(handler HandlerFunc) {
	router.optionsHandler = handler
}

// Running the given context.
func (router *Router) run(ctx *gcontext.GContext) {
	if ctx.GetRequestMethod() == http.MethodOptions {
		if router.optionsHandler != nil {
			router.optionsHandler(ctx)

			return
		}

		return
	}

	// Try to find a matching route, inside the customRoutes.
	// Keep in mind, it may come with queryParams, so we should
	// be matching without it!
	normUrl := ctx.GetUrlWithoutQueryParams()
	if router.mockTree != nil {
		h, _ := router.mockTree.findNode(normUrl)

		// If its not nil, meaning we have to call it.
		if h != nil {
			// The last element of the chain is the handler itself.
			h.mwChain.getLast()(ctx)

			return
		}
	}

	chain, params := router.findRoute(ctx.GetRequestMethod(), normUrl)

	ctx.SetParams(params)

	// If we matched a url.
	if chain != nil {
		chain.run(ctx)

		return
	}

	// Look for service in the serviceEntity, matching the prefix.
	serviceEntity := router.serviceRegisty.FindService(normUrl)

	if serviceEntity != nil {
		// If the service is unavailable, dont try to run it.
		if !serviceEntity.IsAvailable() {
			ctx.SendRaw([]byte{}, http.StatusServiceUnavailable, http.Header{})

			return
		}

		// Get the middlwareChain.
		mwChain := router.getMwChain(ctx.GetUrlParts(), serviceEntity.GetService().Handle)

		// Now execute the middlewareChain.
		mwChain.run(ctx)

		return
	}

	if router.notFoundHandler != nil {
		router.notFoundHandler(ctx)

		return
	}

	// Just in case, send default http 404.
	ctx.SendNotFound()
}

func (router *Router) _poolTester(w http.ResponseWriter, r *http.Request) {
	ctx := router.contextPool.Get().(*gcontext.GContext)
	ctx.Reset(w, r) // Setting the context, to the current params.
	// After the run of the handler, put it back to the pool.
	router.contextPool.Put(ctx)
}

func (router *Router) _createTester(w http.ResponseWriter, r *http.Request) {
	c := gcontext.New(w, r)

	c.SendOk()
}

// Function to implement the http.Handler.
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := router.contextPool.Get().(*gcontext.GContext)
	ctx.Reset(w, r) // Setting the context to the current params.

	fmt.Printf("[%s]: %s\n", ctx.GetRequestMethod(), ctx.GetFullUrl())

	router.run(ctx)

	// After the run of the handler, put it back to the pool.
	router.contextPool.Put(ctx)
}

// Sets the mocks for the router.
func (router *Router) SetMocks(mocks *[]mock.MockCall) {
	tree := createTree()

	// Ranging over the mocks, we create a nodeList
	// and then append to the tree.
	// And finally, setting the tree to the routers
	// current mock tree.
	for _, m := range *mocks {
		node, err := createNodeList(m.Url, createNewMWChain(func(g *gcontext.GContext) {
			// In this scenario, the mock endpoint is only returning
			// some kind of JSON object.
			header := http.Header{}
			header.Add("Content-Type", gcontext.JsonContentType)

			fmt.Printf("MOCK CALL FOR: %s\n", m.Url)
			g.SendRaw(m.Data, m.StatusCode, header)
		}))

		if err != nil {
			fmt.Printf("[ROUTER]: mock create node list error: %v\n", err)

			continue
		}

		if err := tree.addToTree(node); err != nil {
			fmt.Printf("[ROUTER]: mock tree add error: %v\n", err)
		}
	}

	router.mockTree = tree
}

// ----------------------
// |		   Routes			  |
// ----------------------

// Registering @GET method route, with given handlerfunc.
func (r *Router) Get(url string, fn HandlerFunc, mw ...HandlerFunc) error {
	mwChain := createNewMWChain(fn, mw...)

	return r.addRoute(url, http.MethodGet, mwChain)
}

// Registering @POST method route, with given handlefunc.
func (r *Router) Post(url string, fn HandlerFunc, mw ...HandlerFunc) error {
	mwChain := createNewMWChain(fn, mw...)

	return r.addRoute(url, http.MethodPost, mwChain)
}

// Registering the route.
func (r *Router) addRoute(url, method string, mwChain *middlewareChain) error {
	if r == nil {
		return fmt.Errorf("router is nil")
	}

	if !strings.HasPrefix(url, "/api") {
		return errMustStartWithApi
	}

	mTree, exists := r.methodTrees[method]

	// If there is no tree, regarding to
	// the http method, we should create it.
	if !exists {
		mTree = createTree()
		r.methodTrees[method] = mTree
	}

	node, err := createNodeList(url, mwChain)

	if err != nil {
		return err
	}

	err = mTree.addToTree(node)

	if err != nil {
		return err
	}

	return nil
}

// Finds a route by given method and url.
func (r *Router) findRoute(method, url string) (*middlewareChain, map[string]string) {
	tree, exists := r.methodTrees[method]

	if !exists {
		return nil, nil
	}

	node, params := tree.findNode(url)

	if node == nil {
		return nil, nil
	}

	return node.mwChain, params
}

// -------------------
// |   MIDDLEWARES   |
// -------------------

// Creates and returns a pointer to a new Middlware.
func CreateMiddleware(part string, handler HandlerFunc) *Middleware {
	if part == "" {
		return nil
	}

	if handler == nil {
		return nil
	}

	return &Middleware{
		part:    part,
		handler: handler,
	}
}

// Registering a new middleware
func (r *Router) RegisterMiddleware(mw *Middleware) {
	_, exists := r.middlewares[mw.part]

	if exists {
		fmt.Printf("[ROUTER]: middleware (%s) already exists. Ignoring.\n", mw.part)

		return
	}

	r.middlewares[mw.part] = mw.handler
}

// Returns a slice of handleFuncs as middlewares.
func (router *Router) getMwChain(urlParts []string, handler HandlerFunc) *middlewareChain {
	chain := make(map[string]HandlerFunc)

	for _, p := range urlParts {
		handlerF, exists := router.middlewares[p]

		if !exists {
			continue
		}

		if _, contains := chain[p]; contains {
			continue
		}

		chain[p] = handlerF
	}

	handlers := []HandlerFunc{}

	for _, v := range chain {
		handlers = append(handlers, v)
	}

	handlers = append(handlers, handler)

	return &middlewareChain{
		chain: &handlers,
	}
}

// ----------------------
// |		   HELPERS 		  |
// ----------------------

func (r *Router) DisplayRoutes() {
	for k, v := range r.methodTrees {
		fmt.Println(k)
		v.getRoutes()
		fmt.Println("*****")
	}
}
