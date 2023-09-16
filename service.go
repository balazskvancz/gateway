package gateway

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type serviceState uint8

// Since there is more than one supported service types
// – these are for now: REST and gRPC – you can explicitly
// set in the config the appropriate one. If there is no given
// then the fallback is REST.
type serviceType uint8

const (
	StateRegistered serviceState = iota
	StateUnknown
	StateRefused
	StateAvailable

	defaultStatusPath = "/api/status/health-check"
	timeOutSec        = 10

	timeOutDur = timeOutSec * time.Second
)

const (
	serviceRESTType serviceType = iota
	serviceGRPCType
)

var stateTexts = map[serviceState]string{
	StateAvailable:  "available",
	StateRefused:    "refused",
	StateRegistered: "registered",
	StateUnknown:    "unknown",
}

var (
	enabledProtocols      = []string{"http", "https"}
	supportedServiceTypes = []serviceType{serviceRESTType, serviceGRPCType}
)

type ServiceConfig struct {
	// See the type def.
	ServiceType serviceType `json:"serviceType"`

	// The name of the service, must be unique.
	Name string `json:"name"`
	// The unique prefix which identified a URL
	Prefix string `json:"prefix"`

	// Which protocol is used to call the service.
	// Only http and https are supported.
	Protocol string `json:"protocol"`

	// The basic host:port format, where each service listens at.
	Host string `json:"host"`
	Port string `json:"port"`

	// How many seconds it should wait before timeout.
	TimeOutSec int `json:"timeOutSec"`

	// The url to call for healtcheck.
	StatusPath string `json:"statusPath"`
}

type Service interface {
	Handle(*Context)
	Get(string, ...http.Header) (*http.Response, error)
	Post(string, []byte, ...http.Header) (*http.Response, error)
	Put(string, []byte, ...http.Header) (*http.Response, error)
	Delete(string, ...http.Header) (*http.Response, error)
}

type service struct {
	*ServiceConfig

	state      serviceState
	clientPool sync.Pool
}

var _ Service = (*service)(nil)

func newService(conf *ServiceConfig) *service {
	statusPath := func() string {
		if conf != nil && conf.StatusPath != "" {
			return conf.StatusPath
		}
		return defaultStatusPath
	}()

	serv := &service{
		state: StateUnknown,
		ServiceConfig: &ServiceConfig{
			ServiceType: conf.ServiceType,
			Name:        conf.Name,
			Prefix:      conf.Prefix,
			Protocol:    conf.Protocol,
			Host:        conf.Host,
			Port:        conf.Port,
			TimeOutSec:  conf.TimeOutSec,
			StatusPath:  statusPath,
		},
	}

	duration := func() time.Duration {
		if conf != nil && conf.TimeOutSec != 0 {
			return time.Duration(conf.TimeOutSec) * time.Second
		}
		return defaultClientTimeout
	}()

	serv.clientPool = sync.Pool{
		New: func() any {
			return newHttpClient(withHostName(serv.GetAddressWithProtocol()), withTimeOut(duration))
		},
	}

	return serv
}

// validateService validates a service by the given config.
// It returns the first error that occured.
func validateService(config *ServiceConfig) error {
	if config == nil {
		return errConfigIsNil
	}
	if !includes(supportedServiceTypes, config.ServiceType) {
		return errUnsupportedServiceType
	}
	if config.Host == "" {
		return errEmptyHost
	}
	if config.Name == "" {
		return errEmptyName
	}
	if config.Port == "" {
		return errEmptyPort
	}
	if config.Prefix == "" {
		return errEmptyPrefix
	}
	if !includes(enabledProtocols, config.Protocol) {
		return errBadProtocol
	}
	return nil
}

// A handler for each service. For this version it only supports REST based
// handling.
func (s *service) Handle(ctx *Context) {
	if s.ServiceType != serviceRESTType {
		return
	}

	if s.state != StateAvailable {
		ctx.SendUnavailable()

		return
	}

	cl := s.clientPool.Get().(httpClient)
	defer s.clientPool.Put(cl)

	res, err := cl.pipe(ctx.GetRequest())
	if err != nil {
		s.setState(StateRefused)
		// [TODO]: Change it to logger.
		// fmt.Println(err)
		ctx.SendInternalServerError()
		return
	}

	ctx.Pipe(res)
}

// Sending @GET request to the service.
func (s *service) Get(url string, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodGet, url, nil, header...)
}

// Sending @POST request to the service.
func (s *service) Post(url string, data []byte, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodPost, url, bytes.NewReader(data), header...)
}

// Sending @PUT request to the service.
func (s *service) Put(url string, data []byte, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodPut, url, bytes.NewReader(data), header...)
}

// Sending @Delete request to the service.
func (s *service) Delete(url string, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodDelete, url, nil, header...)
}

func (s *service) PostReader(url string, data io.Reader, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodPost, url, data, header...)
}

func (s *service) doRequest(method string, url string, body io.Reader, header ...http.Header) (*http.Response, error) {
	if s.ServiceType != serviceGRPCType {
		return nil, fmt.Errorf("[%s]: is not a REST type service, cant perform HTTP %s", s.Name, method)
	}
	if s.state != StateAvailable {
		return nil, errServiceNotAvailable
	}

	cl := s.clientPool.Get().(httpClient)
	defer s.clientPool.Put(cl)

	return cl.doRequest(method, url, body, header...)
}

// checkStatus checks the status of the service.
func (s *service) checkStatus() error {
	// Little hack for now. We perform the healthcheck only for REST services.
	if s.ServiceType != serviceRESTType {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeOutDur)
	defer cancel()

	cl := s.clientPool.Get().(httpClient)
	defer s.clientPool.Put(cl)

	url := s.GetAddressWithProtocol() + s.StatusPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.setState(StateUnknown)
		return err
	}

	res, err := cl.Do(req)
	if err != nil {
		s.setState(StateRefused)
		return err
	}

	if res.StatusCode != http.StatusOK {
		s.setState(StateRefused)
		return nil
	}

	s.setState(StateAvailable)
	return nil
}

func (s *service) setState(state serviceState) {
	s.state = state
}

func (s *service) GetAddressWithProtocol() string {
	return fmt.Sprintf("%s://%s:%s", s.Protocol, s.Host, s.Port)
}

func (s *service) GetAddress() string {
	return fmt.Sprintf("%s:%s", s.Host, s.Port)
}
