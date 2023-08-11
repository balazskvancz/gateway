package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

type (
	HandlerFunc      func(*Context)
	MiddlewareFunc   func(*Context, HandlerFunc)
	PanicHandlerFunc func(*Context, interface{})

	GatewayOptionFunc func(*Gateway)

	runLevel uint8
)

var (
	defaultContext = context.Background()
)

const (
	defaultAddress = 8000

	routeSystemInfo         = "/api/system/services/info"
	routeUpdateServiceState = "/api/system/services/update"
)

const (
	lvlDev     runLevel = 1 << iota // 1
	lvlProd                         // 2
	mwDisabled                      // 4
	mwEnabled                       // 8

	defaultStartLevel = lvlProd + mwEnabled
)

type GatewayInfo struct {
	// Listening Address for incming HTTP/1* connections.
	address int

	//
	runLevel runLevel

	// The secret key which is used to authenticate amongst services.
	secretKey string

	// The time when the Gateway instance was booted up.
	startTime time.Time

	healthCheckFrequency time.Duration
}

type Gateway struct {
	info *GatewayInfo
	// Base-context of the Gateway.
	ctx context.Context

	// Trees for all the registered endpoints.
	// Every HTTP Method gets a different, by default empty
	// tree, then stored in a map, where the key is the
	// method itself.
	methodTrees map[string]*tree[*Route]

	// The registy which stores all the registered services.
	serviceRegisty *registry

	// A pool for Context.
	contextPool sync.Pool

	// mockTree *tree

	// The registry for all the globally registered middlwares.
	// We store two different types of middlewares.
	// There is one for before all execution and one
	// for after all execution order.
	middlewares []Middleware

	// Custom handler for HTTP 404. Everytime a specific
	// route is not found or a service returned 404 it gets called.
	// By default, there a default notFoundHandler, which sends 404 in header.
	notFoundHandler HandlerFunc

	// Custom handler for HTTP OPTIONS.
	optionsHandler HandlerFunc

	// Custom handler function for panics.
	panicHandler PanicHandlerFunc

	logger logger
}

var _ (http.Handler) = (*Gateway)(nil)

func defaultNotFoundHandler(ctx *Context) {
	ctx.SendNotFound()
}

func defaultPanicHandler(ctx *Context, rec interface{}) {
	errorMsg, ok := rec.(string)
	if !ok {
		return
	}
	ctx.Error(errorMsg)
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

func WithAddress(address int) GatewayOptionFunc {
	return func(g *Gateway) {
		g.info.address = address
	}
}

func WithMiddlewaresEnabled(val runLevel) GatewayOptionFunc {
	var a runLevel = (runLevel)(val+1) << 2

	return func(g *Gateway) {
		g.info.runLevel += a
	}
}

func WithProductionLevel(val runLevel) GatewayOptionFunc {
	var a runLevel = (runLevel)(val + 1)

	return func(g *Gateway) {
		g.info.runLevel += a
	}
}

func WithSecretKey(key string) GatewayOptionFunc {
	return func(g *Gateway) {
		g.info.secretKey = key
	}
}

func WithService(conf *ServiceConfig) GatewayOptionFunc {
	return func(g *Gateway) {
		if err := g.RegisterService(conf); err != nil {
			g.logger.Warning(err.Error())
		}
	}
}

func WithHealthCheckFrequency(t time.Duration) GatewayOptionFunc {
	return func(g *Gateway) {
		g.info.healthCheckFrequency = t
	}
}

// NewFromConfig creates and returns a new Gateway based on
// the given config file path. In case of any errors
// – due to IO reading or marshal error – it returns the error also.
func NewFromConfig(path ...string) (*Gateway, error) {
	finalPath := func() string {
		if len(path) > 0 {
			return path[0]
		}
		return defaultConfigPath
	}()

	opts, err := ReadConfig(finalPath)
	if err != nil {
		return nil, err
	}
	return New(opts...), nil
}

// New returns a new instance of the gateway
// decorated with the given opts.
func New(opts ...GatewayOptionFunc) *Gateway {
	var (
		channel = getContextIdChannel()
		logger  = newGatewayLogger()
	)

	gw := &Gateway{
		info: &GatewayInfo{
			address:              defaultAddress,
			startTime:            time.Now(),
			healthCheckFrequency: defaultHealthCheckFreq,
		},

		ctx:         defaultContext,
		methodTrees: make(map[string]*tree[*Route]),

		serviceRegisty: nil,

		contextPool: sync.Pool{
			New: func() interface{} {
				return newContext(channel, logger)
			},
		},

		middlewares: make([]Middleware, 0),

		notFoundHandler: defaultNotFoundHandler,
		panicHandler:    defaultPanicHandler,
		logger:          logger,
	}

	for _, o := range opts {
		o(gw)
	}

	gw.serviceRegisty = newRegistry(
		withLogger(logger),
		withHealthCheck(gw.info.healthCheckFrequency),
	)

	gw.RegisterMiddleware(
		loggerMiddleware(gw), DefaultMiddlewareMatcher, MiddlewarePostRunner,
	)

	gw.Post(routeSystemInfo, getSystemInfoHandler(gw)).
		registerMiddleware(validateIncomingRequest(gw, func(b []byte) (any, error) { return nil, nil }))

	gw.Post(routeUpdateServiceState, serviceStateUpdateHandler(gw)).
		registerMiddleware(validateIncomingRequest(gw, func(b []byte) (any, error) {
			in := &updateServiceStateRequest{}
			err := json.Unmarshal(b, in)

			return in, err
		}))

	return gw
}

// Start the main process for the Gateway.
// It listens until it receives the signal to close it.
// This method sutable for graceful shutdown.
func (gw *Gateway) Start() {
	if gw.info.runLevel == 0 {
		gw.info.runLevel = defaultStartLevel
	}

	addr := fmt.Sprintf(":%d", gw.info.address)

	gw.logger.Info(
		fmt.Sprintf("The gateway started at %s\tProduction: %t\tMiddlewares enabled: %t",
			addr,
			gw.isProd(),
			gw.areMiddlewaresEnabled()),
	)

	srv := http.Server{
		Addr:    addr,
		Handler: gw,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			gw.logger.Error(fmt.Sprintf("server listen and serve error: %v", err))

			os.Exit(2)
		}

	}()

	// Updating the status of each service.
	go gw.serviceRegisty.updateStatus()

	// Creating a channel, that listens for quiting.
	sigCh := make(chan os.Signal, 1)

	// If there is interrupt by the os, or we manually stop
	// the server, it will notify the created channel,
	// so we can make the shutdown graceful.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh

	gw.logger.clean()

	if err := srv.Shutdown(gw.ctx); err != nil {
		gw.logger.Error(fmt.Sprintf("shutdown err: %v", err))
		os.Exit(1)
	}

	gw.logger.Info("the gateway stopped")
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
func (gw *Gateway) ListenForMocks(_ *[]any) {
	// Just in case, if its not DEV mode
	// we should never update the mocks!
	if gw.isProd() {
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
func (gw *Gateway) RegisterMiddleware(fn MiddlewareFunc, matcher MatcherFunc, mwType ...MiddlewareType) error {
	t := func() MiddlewareType {
		if len(mwType) > 0 {
			return mwType[0]
		}
		return MiddlewarePreRunner
	}()

	mw := newMiddleware(fn,
		withMiddlewareMatcherFunc(matcher),
		withMiddlewareType(t),
	)

	gw.middlewares = append(gw.middlewares, mw)

	return nil
}

func (gw *Gateway) getOrCreateMethodTree(method string) *tree[*Route] {
	tree, exists := gw.methodTrees[method]

	if exists {
		return tree
	}

	t := newTree[*Route]()
	gw.methodTrees[method] = t

	return gw.methodTrees[method]
}

func (gw *Gateway) addRoute(method, url string, handler HandlerFunc) *Route {
	var (
		route = newRoute(url, handler)
		tree  = gw.getOrCreateMethodTree(method)
	)

	if err := tree.insert(url, route); err != nil {
		gw.logger.Warning(fmt.Sprintf("inserting a route with method %s and url %s. Error: %v", method, url, err))
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
	if isNil(node) {
		return nil
	}

	if !node.isLeaf() {
		return nil
	}

	route := node.value

	pathParams := getPathParams(route.fullUrl, url)
	ctx.setParams(pathParams)

	return route
}

// serve serves the context by its HTTP method and URL.
func (gw *Gateway) serve(ctx *Context) {
	// In case of any panics, we catch it and log it.
	defer func() {
		if !gw.isProd() {
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

	var (
		middlewares = gw.filterMatchingMiddlewares(ctx)
		handler     = gw.getMatchingHandlerFunc(ctx)
	)

	finalHandler := middlewares.getHandler(handler)

	finalHandler(ctx)
}

// getMatchingHandlerFunc returns the handler to matches to the given context.
// Firstly it looks amongs the named routes, then among the available services,
// then it returns a 404 handler.
func (gw *Gateway) getMatchingHandlerFunc(ctx *Context) HandlerFunc {
	// Firstly we look among the named routes.
	// If we have some explicit match, then we have to
	// execute its mwchain.
	if route := gw.findNamedRoute(ctx); route != nil {
		if gw.areMiddlewaresEnabled() {
			return route.getChain()
		}
		return route.getHandler()
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

func (g *Gateway) filterMatchingMiddlewares(ctx *Context) *matchingMiddleware {
	mm := &matchingMiddleware{
		pre:  make([]Middleware, 0),
		post: make([]Middleware, 0),
	}

	areEnabled := g.areMiddlewaresEnabled()

	reduce(g.middlewares, func(acc *matchingMiddleware, curr Middleware) *matchingMiddleware {
		if !curr.DoesMatch(ctx) {
			return acc
		}

		if curr.IsPreRunner() {
			if areEnabled {
				acc.pre = append(acc.pre, curr)
			}
			return acc
		}

		acc.post = append(acc.post, curr)
		return acc
	}, mm)

	return mm
}

func writeToResponseMiddleware(ctx *Context) {
	ctx.writer.writeToResponse()
}

// isProd returns whether the the GW is running in production env.
func (g *Gateway) isProd() bool {
	return g.info.runLevel&lvlProd != 0
}

// areMiddlewaresEnabled returns whether the the middlewares are enabled.
func (g *Gateway) areMiddlewaresEnabled() bool {
	return g.info.runLevel&mwEnabled != 0
}

func loggerMiddleware(g *Gateway) MiddlewareFunc {
	return func(ctx *Context, next HandlerFunc) {
		g.logger.Info(string(ctx.getLog()))
		next(ctx)
	}
}

// RegisterService creates and registers a new Service to the registry
// based on the given config. In case of validation error or duplicate
// service, it returns error.
func (g *Gateway) RegisterService(conf *ServiceConfig) error {
	return g.serviceRegisty.addService(conf)
}
