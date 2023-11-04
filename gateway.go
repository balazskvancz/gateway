package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/balazskvancz/gorouter"
)

type (
	GatewayOptionFunc = func(*Gateway)

	Context        = gorouter.Context
	ContextKey     = gorouter.ContextKey
	HandlerFunc    = gorouter.HandlerFunc
	MiddlewareFunc = gorouter.MiddlewareFunc
	Middleware     = gorouter.Middleware
	Route          = gorouter.Route

	runLevel = uint8
)

var (
	defaultContext = context.Background()
)

const (
	defaultAddress = 8000

	routeSystemPrefix = "/api/system"

	routeSystemInfo         = routeSystemPrefix + "/services/info"
	routeUpdateServiceState = routeSystemPrefix + "/services/update"
)

const (
	lvlDev     runLevel = 1 << iota // 1
	lvlProd                         // 2
	mwDisabled                      // 4
	mwEnabled                       // 8

	defaultStartLevel = lvlProd + mwEnabled

	JsonContentType     = "application/json"
	JsonContentTypeUTF8 = JsonContentType + "; charset=UTF-8"
	TextHtmlContentType = "text/html"
	XmlContentType      = "application/xml"
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

	grpcProxyAddress int
}

type Gateway struct {
	info *GatewayInfo

	router gorouter.Router

	// Base-context of the Gateway.
	ctx context.Context

	// The registy which stores all the registered services.
	serviceRegisty *registry

	// TODO: make mockTree for development purposes.
	// mockTree *tree

	// Custom handler for HTTP 404. Everytime a specific
	// route is not found or a service returned 404 it gets called.
	// By default, there a default notFoundHandler, which sends 404 in header.
	notFoundHandler HandlerFunc

	// Custom handler function for panics.
	panicHandler gorouter.PanicHandlerFunc

	grpcProxy *grpcProxy

	logger logger
}

func defaultNotFoundHandler(ctx Context) {
	ctx.SendNotFound()
}

func defaultPanicHandler(ctx Context, rec interface{}) {
	errorMsg, ok := rec.(string)
	if !ok {
		return
	}
	ctx.Error(errorMsg)
	ctx.SendInternalServerError()
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

func WithDisabledLoggers(disabled logTypeValue) GatewayOptionFunc {
	return func(g *Gateway) {
		g.logger.disable(disabled)
	}
}

func WithGrpcProxy(addr int) GatewayOptionFunc {
	return func(g *Gateway) {
		g.info.grpcProxyAddress = addr
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
		// channel = getContextIdChannel()
		logger = newGatewayLogger()
	)

	gw := &Gateway{
		info: &GatewayInfo{
			address:              defaultAddress,
			startTime:            time.Now(),
			healthCheckFrequency: defaultHealthCheckFreq,
		},

		ctx: defaultContext,

		serviceRegisty: newRegistry(),

		notFoundHandler: defaultNotFoundHandler,
		panicHandler:    defaultPanicHandler,
		logger:          logger,
	}

	for _, o := range opts {
		o(gw)
	}

	// Initiating the router entity.
	gw.router = gorouter.New(
		gorouter.WithAddress(gw.info.address),
		gorouter.WithServerName(fmt.Sprintf("api-gateway %s / goRouter", Version)),
		gorouter.WithNotFoundHandler(gw.serve),
		gorouter.WithEmptyTreeHandler(gw.serve),
		gorouter.WithMiddlewaresEnabled(gw.areMiddlewaresEnabled()),
	)

	gw.serviceRegisty.withHealthCheck(gw.info.healthCheckFrequency)
	gw.serviceRegisty.withLogger(gw.logger)

	// If there was a gRPC address given via config, then attach the proxy.
	if gw.info.grpcProxyAddress != 0 {
		gw.grpcProxy = newGrpcProxy(gw.info.grpcProxyAddress, gw.logger, gw.getGRPCServiceByPrefix)
	}

	gw.registerSystemRoutes()

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

	// If there is a gRPC proxy attached to the Gateway
	// then it should start listening.
	if gw.grpcProxy != nil {
		go gw.grpcProxy.listen()

		defer gw.grpcProxy.stop()
	}

	// Updating the status of each service.
	go gw.serviceRegisty.updateStatus()

	ctx, cancel := context.WithCancel(context.Background())

	go gw.router.ListenWithContext(ctx)

	// Creating a channel, that listens for quiting.
	sigCh := make(chan os.Signal, 1)

	// If there is interrupt by the os, or we manually stop
	// the server, it will notify the created channel,
	// so we can make the shutdown graceful.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh

	cancel()
	gw.logger.clean()

	gw.logger.Info("the gateway stopped")
}

// GetService searches for a service by its name.
// Returns error if, there is no service by the given name.
func (gw *Gateway) GetService(name string) (Service, error) {
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
func (gw *Gateway) Get(url string, handler HandlerFunc) Route {
	return gw.router.Get(url, handler)
}

// Post registers a custom route with method @POST.
func (gw *Gateway) Post(url string, handler HandlerFunc) Route {
	return gw.router.Post(url, handler)
}

// Put registers a custom route with method @PUT.
func (gw *Gateway) Put(url string, handler HandlerFunc) Route {
	return gw.router.Put(url, handler)
}

// Delete registers a custom route with method @DELETE.
func (gw *Gateway) Delete(url string, handler HandlerFunc) Route {
	return gw.router.Delete(url, handler)
}

// Head registers a custom route with method @HEAD.
func (gw *Gateway) Head(url string, handler HandlerFunc) Route {
	return gw.router.Head(url, handler)
}

// RegisterMiddleware registers a middleware instance to the gateway.
func (gw *Gateway) RegisterMiddleware(mw ...Middleware) error {
	gw.router.RegisterMiddlewares(mw...)

	return nil
}

func (gw *Gateway) serve(ctx Context) {
	s := gw.serviceRegisty.findService(ctx.GetCleanedUrl())
	if s != nil {
		s.Handle(ctx)

		return
	}

	// Otherwise it is the default 404 handler.
	ctx.SendNotFound()
}

// isProd returns whether the the GW is running in production env.
func (g *Gateway) isProd() bool {
	return g.info.runLevel&lvlProd != 0
}

// areMiddlewaresEnabled returns whether the the middlewares are enabled.
func (g *Gateway) areMiddlewaresEnabled() bool {
	return g.info.runLevel&mwEnabled != 0
}

// RegisterService creates and registers a new Service to the registry
// based on the given config. In case of validation error or duplicate
// service, it returns error.
func (g *Gateway) RegisterService(conf *ServiceConfig) error {
	return g.serviceRegisty.addService(conf)
}

func (g *Gateway) getGRPCServiceByPrefix(p string) Service {
	if p == "" {
		return nil
	}
	serv := g.serviceRegisty.findService(p)
	if serv == nil {
		return nil
	}
	if serv.ServiceType != serviceGRPCType {
		return nil
	}
	return serv
}

func (gw *Gateway) registerSystemRoutes() {
	systemMatcher := func(ctx Context) bool {
		return strings.HasPrefix(ctx.GetUrl(), routeSystemPrefix)
	}

	mwFunc := func(ctx Context, next HandlerFunc) {
		u := ctx.GetCleanedUrl()

		if u == routeSystemInfo {
			fn := validateIncomingRequest(gw,
				func(b []byte) (any, error) { return nil, nil },
			)

			fn(ctx, next)

			return
		}

		if u == routeUpdateServiceState {
			fn := validateIncomingRequest(gw, func(b []byte) (any, error) {
				var (
					in  = &updateServiceStateRequest{}
					err = json.Unmarshal(b, in)
				)

				return in, err
			})

			fn(ctx, next)

			return
		}
	}

	mw := gorouter.NewMiddleware(
		mwFunc,
		gorouter.MiddlewareWithMatchers(systemMatcher),
		gorouter.MiddlewareWithAlwaysAllowed(true),
	)

	var _ = mw

	gw.RegisterMiddleware(mw)

	gw.Post(routeSystemInfo, getSystemInfoHandler(gw))
	gw.Post(routeUpdateServiceState, serviceStateUpdateHandler(gw))
}
