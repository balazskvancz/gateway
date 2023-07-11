package gateway

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/balazskvancz/gateway/pkg/communicator"
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

type ServiceConfig struct {
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	Prefix   string `json:"prefix"`
}

type Service interface {
	Handle(*Context)
}

type service struct {
	*ServiceConfig
	client *communicator.HttpClient

	state      serviceState
	clientPool sync.Pool
}

var _ Service = (*service)(nil)

type services = *[]service

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

// Validates the given services. Returns error if
// something is not correct.
func validateServices(configs []*ServiceConfig) error {
	if len(configs) == 0 {
		return errServicesSliceIsEmpty
	}

	//
	for _, service := range configs {
		if len(service.Prefix) == 0 {
			return errNoService
		}
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
		ctx.SendInternalServerError()
		return
	}

	ctx.Pipe(res)
}

// Sending @GET request to the service.
func (s *service) Get(url string, header ...http.Header) (*http.Response, error) {
	if s.state != StateAvailable {
		return nil, errServiceNotAvailable
	}

	// If there is any given
	if len(header) > 0 {
		s.client.SetHeader(header[0])
	}

	ch := s.client.PoolGet()
	defer s.client.PoolPut(ch)

	res, err := s.client.Get(url)

	return res, err
}

// Sending @POST request to the service.
func (s *service) Post(url string, data []byte, header ...http.Header) (*http.Response, error) {
	if s.state != StateAvailable {
		return nil, errServiceNotAvailable
	}

	if len(header) > 0 {
		s.client.SetHeader(header[0])
	}

	ch := s.client.PoolGet()
	defer s.client.PoolPut(ch)

	res, err := s.client.Post(url, data)

	return res, err
}

// checkStatus checks the status of the service.
func (s *service) checkStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutDur)
	defer cancel()

	url := s.GetAddress() + statusPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.setState(StateUnknown)
		return err
	}

	res, err := s.client.Do(req)
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

func (s *service) CreateClient() {
	s.client = communicator.New(s.GetAddress(), timeOutDur)
}
