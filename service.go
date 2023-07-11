package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

type Service struct {
	*ServiceConfig
	client *communicator.HttpClient

	state serviceState
}

type services = *[]Service

// Validates the given services. Returns error if
// something is not correct.
func ValidateServices(srvcs services) error {
	if srvcs == nil {
		return errServicesIsNil
	}

	if len(*srvcs) == 0 {
		return errServicesSliceIsEmpty
	}

	// Rule.
	// Every services prefix must be have length and the parts must be equal.
	prefixLength := len(strings.Split((*srvcs)[0].Prefix, "/"))

	isZeroLengthOk, isSamePrefixOk := true, true

	//
	for _, service := range *srvcs {
		if len(service.Prefix) == 0 {
			isZeroLengthOk = false
		}

		if len(strings.Split(service.Prefix, "/")) != prefixLength {
			isSamePrefixOk = false
		}
	}

	if !isZeroLengthOk {
		return errServicesPrefixLength
	}

	if !isSamePrefixOk {
		return errServicesSamePrefixLength
	}

	return nil
}

// A handler for each service.
func (s *Service) Handle(ctx *Context) {
	if s.state != StateAvailable {
		ctx.SendUnavailable()

		return
	}

	// Forwarding the data to the the service.
	b, code, header := s.client.Forwarder(ctx.GetRequest())

	ctx.SendRaw(b, code, header)
}

// Sending @GET request to the service.
func (s *Service) Get(url string, header ...http.Header) (*http.Response, error) {
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
func (s *Service) Post(url string, data []byte, header ...http.Header) (*http.Response, error) {
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

// Checks the status of the service.
func (s *Service) CheckStatus() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeOutDur)
	defer cancel()

	url := s.GetAddress() + statusPath

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		s.state = StateUnknown
		return false, err
	}

	res, err := s.client.Do(req)
	if err != nil {
		s.state = StateRefused
		return false, err
	}

	if res.StatusCode != http.StatusOK {
		s.state = StateRefused
		return false, errServiceNotAvailable
	}

	s.state = StateAvailable
	return true, nil
}

func (s *Service) GetAddress() string {
	return fmt.Sprintf("%s://%s:%s", s.Protocol, s.Host, s.Port)
}

func (s *Service) CreateClient() {
	s.client = communicator.New(s.GetAddress(), timeOutDur)
}
