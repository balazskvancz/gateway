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

const (
	StateRegistered serviceState = iota
	StateUnknown
	StateRefused
	StateAvailable

	statusPath = "/api/status/health-check"
	timeOutSec = 10

	timeOutDur = timeOutSec * time.Second
)

var enabledProtocols = []string{"http", "https"}

type ServiceConfig struct {
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Prefix   string `json:"prefix"`
}

type Service interface {
	Handle(*Context)
	Get(string, ...http.Header) (*http.Response, error)
	Post(string, []byte, ...http.Header) (*http.Response, error)
}

type service struct {
	*ServiceConfig

	state      serviceState
	clientPool sync.Pool
}

var _ Service = (*service)(nil)

func newService(conf *ServiceConfig) *service {
	serv := &service{
		state:         StateUnknown,
		ServiceConfig: conf,
	}

	serv.clientPool = sync.Pool{
		New: func() any {
			return newHttpClient(withHostName(serv.GetAddress()))
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

// A handler for each service.
func (s *service) Handle(ctx *Context) {
	if s.state != StateAvailable {
		ctx.SendUnavailable()

		return
	}

	cl := s.clientPool.Get().(*httpClient)
	defer s.clientPool.Put(cl)

	res, err := cl.pipe(ctx.GetRequest())
	if err != nil {
		// todo
		fmt.Println(err)
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

func (s *service) PostReader(url string, data io.Reader, header ...http.Header) (*http.Response, error) {
	return s.doRequest(http.MethodPost, url, data, header...)
}

func (s *service) doRequest(method string, url string, body io.Reader, header ...http.Header) (*http.Response, error) {
	if s.state != StateAvailable {
		return nil, errServiceNotAvailable
	}

	cl := s.clientPool.Get().(*httpClient)
	defer s.clientPool.Put(cl)

	return cl.doRequest(method, url, body, header...)
}

// checkStatus checks the status of the service.
func (s *service) checkStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutDur)
	defer cancel()

	cl := s.clientPool.Get().(*httpClient)
	defer s.clientPool.Put(cl)

	url := s.GetAddress() + statusPath

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

func (s *service) GetAddress() string {
	return fmt.Sprintf("%s://%s:%s", s.Protocol, s.Host, s.Port)
}
