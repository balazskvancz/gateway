package gateway

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/balazskvancz/gateway/pkg/config"
	"github.com/balazskvancz/gateway/pkg/mock"
	"github.com/balazskvancz/gateway/pkg/registry"
	"github.com/balazskvancz/gateway/pkg/router"
	"github.com/balazskvancz/gateway/pkg/service"
)

type Gateway struct {
	address         int
	mux             *router.Router
	serviceRegistry *registry.Registry

	isProd bool
}

// Returns a new instance of the gateway.
func New() (*Gateway, error) {
	// Read config.
	cfg, err := config.GetConfig()

	if err != nil {
		return nil, err
	}

	err = service.ValidateServices(cfg.Services)

	if err != nil {
		return nil, err
	}

	registry, err := registry.NewRegistry(cfg.Services, cfg.SleepMin)

	if err != nil {
		return nil, err
	}

	return &Gateway{
		address:         cfg.Address,
		serviceRegistry: registry,
		mux:             router.New(registry),
		isProd:          cfg.IsProd,
	}, nil
}

// Start the main process for the Gateway.
// It listens until it receives the signal to close it.
// This method sutable for graceful shutdown.
func (gw *Gateway) Start() {
	addr := fmt.Sprintf(":%d", gw.address)

	mode := "DEV"

	if gw.isProd {
		mode = "PROD"
	}

	fmt.Printf("The gateway started at %s, in mode: %s\n", addr, mode)

	srv := http.Server{
		Addr:    addr,
		Handler: gw.mux,
	}

	// gw.mux.DisplayRoutes()
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
	go gw.serviceRegistry.UpdateStatus()

	// Creating a channel, that listens for quiting.
	sigCh := make(chan os.Signal, 1)

	// If there is interrupt by the os, or we manually stop
	// the server, it will notify the created channel,
	// so we can make the shutdown graceful.
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	<-sigCh
	if err := srv.Shutdown(context.Background()); err != nil {
		fmt.Printf("[GATEWAY]: shutdown err: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("[GATEWAY]: the server stopped.")
}

// Gets service by its name. Returns error if, there is
// no service by the given name. Also returns error
// if the certain service is not available the that time.
func (gw *Gateway) GetService(name string) (*service.Service, error) {
	servEntity := gw.serviceRegistry.GetServiceByName(name)

	if servEntity == nil {
		return nil, nil
	}

	srvc := servEntity.GetService()

	if !srvc.IsAvailable {
		return nil, service.ErrServiceNotAvailable
	}

	return srvc, nil
}

// Listener for mocks.
func (gw *Gateway) ListenForMocks(mocks *[]mock.MockCall) {
	// Just in case, if its not DEV mode
	// we should never update the mocks!
	if gw.isProd {
		return
	}

	gw.mux.SetMocks(mocks)
}

// -----------------
// | CUSTOM ROUTES |
// -----------------

// Register a custom route with method @GET.
func (gw *Gateway) Get(url string, handler router.HandlerFunc, mw ...router.HandlerFunc) {
	if err := gw.mux.Get(url, handler, mw...); err != nil {
		fmt.Printf("err :%v\n", err)
	}
}

// Register a custom route with method @GET.
func (gw *Gateway) Post(url string, handler router.HandlerFunc, mw ...router.HandlerFunc) {
	gw.mux.Post(url, handler, mw...)
}

// -----------------
// |  MIDDLEWARES  |
// -----------------

// Registering Middleware to the router.
func (gw *Gateway) RegisterMiddleware(part string, handler router.HandlerFunc) {
	mw := router.CreateMiddleware(part, handler)

	if mw == nil {
		return
	}

	gw.mux.RegisterMiddleware(mw)
}
