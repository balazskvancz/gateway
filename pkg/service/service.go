package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/balazskvancz/gateway/pkg/communicator"
	"github.com/balazskvancz/gateway/pkg/gcontext"
)

const (
	statusPath = "/api/status/health-check"
	timeOutSec = 10

	timeOutDur = timeOutSec * time.Second
)

var (
	errServicesIsNil            = errors.New("services is nil")
	errServicesPrefixLength     = errors.New("service prefix must be greater than zero")
	errServicesSamePrefixLength = errors.New("service prefix must be same length")
	errServicesSliceIsEmpty     = errors.New("services slice is empty")

	ErrServiceNotAvailable = errors.New("service not available")
)

type Service struct {
	Protocol     string `json:"protocol"`
	Name         string `json:"name"`
	Host         string `json:"host"`
	Port         string `json:"port"`
	Prefix       string `json:"prefix"`
	CContentType string `json:"ctype"` // Content-type of communication.

	client *communicator.HttpClient

	IsAvailable bool
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
func (s *Service) Handle(ctx *gcontext.GContext) {
	if !s.IsAvailable {
		ctx.SendUnavailable()

		return
	}

	// Forwarding the data to the the service.
	b, code, header := s.client.Forwarder(ctx.GetRequest())

	ctx.SendRaw(b, code, header)
}

// Sending @GET request to the service.
func (s *Service) Get(url string, header ...http.Header) (*http.Response, error) {
	if !s.IsAvailable {
		return nil, ErrServiceNotAvailable
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
	if !s.IsAvailable {
		return nil, ErrServiceNotAvailable
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
		s.IsAvailable = false
		return false, err
	}

	res, err := s.client.Do(req)

	if err != nil {
		s.IsAvailable = false
		return false, err
	}

	if res.StatusCode != http.StatusOK {
		s.IsAvailable = false
		return false, ErrServiceNotAvailable
	}

	s.IsAvailable = true
	return true, nil
}

func (s *Service) GetAddress() string {
	return fmt.Sprintf("%s://%s:%s", s.Protocol, s.Host, s.Port)
}

func (s *Service) CreateClient() {
	s.client = communicator.New(s.GetAddress(), timeOutDur)
}
