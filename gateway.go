package gateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/balazskvancz/gateway/pkg/mock"
)

type (
	HandlerFunc      func(*Context)
	MiddlewareFunc   func(*Context, HandlerFunc)
	PanicHandlerFunc func(*Context, interface{})

	GatewayOptionFunc func(*Gateway)
)

const (
	defaultAddress = 8000
)

var (
	defaultContext = context.Background()
)

type Gateway struct {
	// Listening address for incming HTTP/1* connections.
	address int

	//
	isProd bool

	// Base-context of the Gateway.
	ctx context.Context

	// Trees for all the registered endpoints.
	// Every HTTP Method gets a different, by default empty
	// tree, then stored in a map, where the key is the
	// method itself.
	methodTrees map[string]*tree

	// The registy which stores all the registered services.
	serviceRegisty *registry

	// A pool for Context.
	contextPool sync.Pool

	// mockTree *tree

	// The registry for all the globally registered middlwares.
	// We store two different types of middlewares.
	// There is one for before all execution and one
	// for after all execution order.
	mwRegistry middlewareRegistry

	// Custom handler for HTTP 404. Everytime a specific
	// route is not found or a service returned 404 it gets called.
	// By default, there a default notFoundHandler, which sends 404 in header.
	notFoundHandler HandlerFunc

	// Custom handler for HTTP OPTIONS.
	optionsHandler HandlerFunc

	// Custom handler function for panics.
	panicHandler PanicHandlerFunc
}

var _ (http.Handler) = (*Gateway)(nil)

func defaultNotFoundHandler(ctx *Context) {
	ctx.SendNotFound()
}

func defaultPanicHandler(ctx *Context, rec interface{}) {
	// TODO: logging
	fmt.Println(rec)
	ctx.SendInternalServerError()
}

func getContextIdChannel() contextIdChan {
	ch := make(chan uint64)

	go func() {
		var counter uint64 = 1
		for {
			ch <- counter
			counter++
		}
	}()

	return ch
}

// New returns a new instance of the gateway
// decorated with the given opts.
func New(opts ...GatewayOptionFunc) *Gateway {
	gw := &Gateway{
		address:     defaultAddress,
		ctx:         defaultContext,
		methodTrees: make(map[string]*tree),

		serviceRegisty: newRegistry(withHealthCheck(defaultHealthCheckFreq)),

		contextPool: sync.Pool{
			New: func() interface{} {
				return newContext(getContextIdChannel())
			},
		},

		mwRegistry: *newMiddlewareRegistry(),

		notFoundHandler: defaultNotFoundHandler,
		panicHandler:    defaultPanicHandler,
	}

	for _, o := range opts {
		o(gw)
	}

	return gw
}

// ReadConfig reads the JSON config from given path,
// then returns it as a slice of GatewayOptionFunc,
// which can be passed into the New factory.
// In case of unexpected behaviour, it returns error.
func ReadConfig(path string) ([]GatewayOptionFunc, error) {
	return nil, nil
}

// Start the main process for the Gateway.
// It listens until it receives the signal to close it.
// This method sutable for graceful shutdown.
func (gw *Gateway) Start() {
	addr := fmt.Sprintf(":%d", gw.address)

	mode := func() string {
		if gw.isProd {
			return "PROD"
		}
		return "DEV"
	}()

	// Change to logger.
	fmt.Printf("The gateway started at %s, in mode: %s\n", addr, mode)

	srv := http.Server{
		Addr:    addr,
		Handler: gw,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Printf("server err: %v\n", err)

			os.Exit(2)
		}

	}()

	// If we running inside DEV mode, we use mock calls.
	if !gw.isProd {
		mock := mock.New(gw)

		// If the mock is not nil, we should watch for file change.
		if mock != nil {
			go mock.WatchReload()
		}
	}

	// Updating the status of each service.
	go gw.serviceRegisty.updateStatus()

	// Creating a channel, that listens for quiting.
	sigCh := make(chan os.Signal, 1)

	// If there is interrupt by the os, or we manually stop
	// the server, it will notify the created channel,
	// so we can make the shutdown graceful.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	if err := srv.Shutdown(gw.ctx); err != nil {
		fmt.Printf("[GATEWAY]: shutdown err: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[GATEWAY]: the server stopped.")
}

// GetService searches for a service by its name.
// Returns error if, there is no service by the given name.
func (gw *Gateway) GetService(name string) (*service, error) {
	service := gw.serviceRegisty.getServiceByName(name)
	if service == nil {
		return nil, ErrServiceNotExists
	}
	return service, nil
}

// Listener for mocks.
func (gw *Gateway) ListenForMocks(mocks *[]mock.MockCall) {
	// Just in case, if its not DEV mode
	// we should never update the mocks!
	if gw.isProd {
		return
	}

	//u gw.mux.SetMocks(mocks)
}

// Get registers a custom route with method @GET.
func (gw *Gateway) Get(url string, handler HandlerFunc) *Route {
	return gw.addRoute(http.MethodGet, url, handler)
}

// Post registers a custom route with method @POST.
func (gw *Gateway) Post(url string, handler HandlerFunc) *Route {
	return gw.addRoute(http.MethodPost, url, handler)
}

// Put registers a custom route with method @PUT.
func (gw *Gateway) Put(url string, handler HandlerFunc) *Route {
	return gw.addRoute(http.MethodPut, url, handler)
}

// Delete registers a custom route with method @DELETE.
func (gw *Gateway) Delete(url string, handler HandlerFunc) *Route {
	return gw.addRoute(http.MethodDelete, url, handler)
}

// Head registers a custom route with method @HEAD.
func (gw *Gateway) Head(url string, handler HandlerFunc) *Route {
	return gw.addRoute(http.MethodHead, url, handler)
}

// RegisterMiddleware registers a middleware instance to the gateway.
func (gw *Gateway) RegisterMiddleware(fn MiddlewareFunc, matcher ...matcherFunc) error {
	mw := newMiddleware(fn, matcher...)

	gw.mwRegistry.push(mwPreRunner, mw)

	return nil
}

func (gw *Gateway) getOrCreateMethodTree(method string) *tree {
	tree, exists := gw.methodTrees[method]

	if exists {
		return tree
	}

	t := newTree()
	gw.methodTrees[method] = t

	return gw.methodTrees[method]
}

func (gw *Gateway) addRoute(method, url string, handler HandlerFunc) *Route {
	route := newRoute(url, handler)

	tree := gw.getOrCreateMethodTree(method)

	if err := tree.insert(url, route); err != nil {
		// todo logging
		return nil
	}

	return route
}

func (gw *Gateway) findNamedRoute(ctx *Context) *Route {
	tree, exists := gw.methodTrees[ctx.GetRequestMethod()]
	if !exists {
		return nil
	}

	url := ctx.GetUrlWithoutQueryParams()

	node := tree.find(url)
	if node == nil {
		return nil
	}

	route, ok := node.value.(*Route)
	if !ok {
		return nil
	}

	pathParams := getPathParams(route.fullUrl, url)
	ctx.setParams(pathParams)

	return route
}

// serve serves the context by its HTTP method and URL.
func (gw *Gateway) serve(ctx *Context) {
	// In case of any panics, we catch it and log it.
	defer func() {
		if !gw.isProd {
			return
		}

		prec := recover()

		if prec != nil && gw.panicHandler != nil {
			gw.panicHandler(ctx, prec)
		}
	}()

	// In case of HTTP Options.
	if ctx.GetRequestMethod() == http.MethodOptions {
		optHandler := gw.optionsHandler
		if optHandler != nil {
			optHandler(ctx)

			return
		}

		return
	}

	globalMws := gw.filterMatchingMiddlewares(ctx)

	handler := gw.getMatchingHandlerFunc(ctx)

	finalChain := globalMws.getHandlerFuncSlice(handler)

	finalChain[0](ctx)
}

// getMatchingHandlerFunc returns the handler to matches to the given context.
// Firstly it looks amongs the named routes, then among the available services,
// then it returns a 404 handler.
func (gw *Gateway) getMatchingHandlerFunc(ctx *Context) HandlerFunc {
	// Firstly we look among the named routes.
	// If we have some explicit match, then we have to
	// execute its mwchain.
	if route := gw.findNamedRoute(ctx); route != nil {
		return route.getChain()
	}

	// After try to forward it to specific service.
	s := gw.serviceRegisty.findService(ctx.GetUrlWithoutQueryParams())

	if s != nil {
		return s.Handle
	}

	return func(ctx *Context) {
		ctx.SendNotFound()
	}
}

// ServeHTTP serves the incoming HTTP request.
func (gw *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get a context out of the pool.
	ctx := gw.contextPool.Get().(*Context)
	ctx.reset(w, r)

	// Execute the main logic.
	gw.serve(ctx)

	// Release every pointer then put it back to the pool.
	// If we didnt release the all the pointers, then the GC
	// cant free the pointer until we call ctx.reset on
	// the same pointer.
	ctx.empty()
	gw.contextPool.Put(ctx)
}

func (g *Gateway) filterMatchingMiddlewares(ctx *Context) middlewareChain {
	preMws := g.mwRegistry.get(mwPreRunner)

	var filterFn = func(m *middleware) bool {
		return m.doesMatch(ctx)
	}

	chain := filter(preMws, filterFn)

	return chain
}
