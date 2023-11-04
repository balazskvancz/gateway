package gateway

import (
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
)

func TestValidateService(t *testing.T) {
	type testCase struct {
		name string
		conf *ServiceConfig
		err  error
	}

	tt := []testCase{
		{
			name: "the function returns error if the config ptr is nil",
			conf: nil,
			err:  errConfigIsNil,
		},
		{
			name: "the function returns error if the host is empty",
			conf: &ServiceConfig{},
			err:  errEmptyHost,
		},
		{
			name: "the function returns error if the name is empty",
			conf: &ServiceConfig{
				Host: "mock-host",
			},
			err: errEmptyName,
		},
		{
			name: "the function returns error if the port is empty",
			conf: &ServiceConfig{
				Host: "mock-host",
				Name: "mock-name",
			},
			err: errEmptyPort,
		},
		{
			name: "the function returns error if the prefix is empty",
			conf: &ServiceConfig{
				Host: "mock-host",
				Name: "mock-name",
				Port: "8000",
			},
			err: errEmptyPrefix,
		},
		{
			name: "the function returns error if the protocol is empty",
			conf: &ServiceConfig{
				Host:   "mock-host",
				Name:   "mock-name",
				Port:   "8000",
				Prefix: "/mock",
			},
			err: errBadProtocol,
		},
		{
			name: "the function returns error if the protocol is not supported",
			conf: &ServiceConfig{
				Host:     "mock-host",
				Name:     "mock-name",
				Port:     "8000",
				Prefix:   "/mock",
				Protocol: "foo",
			},
			err: errBadProtocol,
		},
		{
			name: "the function returns nil is the config is valid",
			conf: &ServiceConfig{
				Host:     "mock-host",
				Name:     "mock-name",
				Port:     "8000",
				Prefix:   "/mock",
				Protocol: "http",
			},
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateService(tc.conf); !errors.Is(err, tc.err) {
				t.Errorf("expected error: %v; got error: %v\n", tc.err, err)
			}
		})
	}
}

type serviceFactory func(*testing.T) *service
type mockHttpClient struct {
	mockPipe func(*http.Request) (*http.Response, error)
	mockDo   func(*http.Request) (*http.Response, error)
	// httpClient
}

func (mc *mockHttpClient) pipe(req *http.Request) (*http.Response, error) {
	return mc.mockPipe(req)
}

func (mc *mockHttpClient) Do(req *http.Request) (*http.Response, error) {
	return mc.mockDo(req)
}

func (mc *mockHttpClient) doRequest(_ string, _ string, _ io.Reader, _ ...http.Header) (*http.Response, error) {
	return nil, nil
}

var _ (httpClient) = (*mockHttpClient)(nil)

// For now, these tests are commented out, will fix it later.
/*
func TestHandle(t *testing.T) {
	type testCase struct {
		name       string
		getService serviceFactory

		expStatus int
		expBody   []byte
		expState  serviceState
	}

	restServiceConfig := &ServiceConfig{
		ServiceType: 0,
	}

	tt := []testCase{
		{
			name: "the function send HTTP 503 if the service is not available",
			getService: func(t *testing.T) *service {
				return newService(restServiceConfig)
			},
			expStatus: http.StatusServiceUnavailable,
			expBody:   nil,
			expState:  StateUnknown,
		},
		{
			name: "the function send HTTP 500 if the pipe is not successful",
			getService: func(t *testing.T) *service {
				s := newService(restServiceConfig)

				s.setState(StateAvailable)

				s.clientPool = sync.Pool{
					New: func() any {
						return &mockHttpClient{
							mockPipe: func(r *http.Request) (*http.Response, error) {
								return nil, errors.New("mock-err")
							},
						}
					},
				}

				return s
			},
			expStatus: http.StatusInternalServerError,
			expBody:   nil,
			expState:  StateRefused,
		},
		{
			name: "the function writes the body and statusCode of response from service",
			getService: func(t *testing.T) *service {
				s := newService(restServiceConfig)

				s.setState(StateAvailable)

				s.clientPool = sync.Pool{
					New: func() any {
						return &mockHttpClient{
							mockPipe: func(r *http.Request) (*http.Response, error) {
								res := http.Response{}

								res.StatusCode = http.StatusOK
								res.Body = io.NopCloser(bytes.NewReader([]byte(`{"message":"ok"}`)))

								return &res, nil
							},
						}
					},
				}

				return s
			},
			expStatus: http.StatusOK,
			expBody:   []byte(`{"message":"ok"}`),
			expState:  StateAvailable,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				service = tc.getService(t)
				ctx     = gorouter.NewContext(gorouter.ContextConfig{})
			)

			service.Handle(ctx)

			if ctx != tc.expStatus {
				t.Errorf("expected statusCode: %d; got statusCode: %d\n", tc.expStatus, ctx.writer.statusCode)
			}

			if !reflect.DeepEqual(ctx.writer.b, tc.expBody) {
				t.Errorf("expected body: %s; got body: %s\n", tc.expBody, ctx.writer.b)
			}

			if service.state != tc.expState {
				t.Errorf("expected state: %d; got state: %d\n", tc.expState, service.state)
			}
		})
	}
}
*/

func TestChechStatus(t *testing.T) {
	type testCase struct {
		name       string
		getService serviceFactory

		expErr   error
		expState serviceState
	}

	var (
		httpDoError = errors.New("mock-do-error")
	)

	tt := []testCase{
		{
			name: "the function returns an error the http call returns error",
			getService: func(t *testing.T) *service {
				s := newService(&ServiceConfig{
					Protocol: "http",
					Host:     "localhost",
					Port:     "8000",
				})

				s.clientPool = sync.Pool{
					New: func() any {
						return &mockHttpClient{
							mockDo: func(r *http.Request) (*http.Response, error) {
								return nil, httpDoError
							},
						}
					},
				}

				return s
			},
			expErr:   httpDoError,
			expState: StateRefused,
		},
		{
			name: "the function returns no error and sets the state to `StateRefused`",
			getService: func(t *testing.T) *service {
				s := newService(&ServiceConfig{
					Protocol: "http",
					Host:     "localhost",
					Port:     "8000",
				})

				s.clientPool = sync.Pool{
					New: func() any {
						return &mockHttpClient{
							mockDo: func(r *http.Request) (*http.Response, error) {
								res := &http.Response{}

								res.StatusCode = http.StatusBadRequest

								return res, nil
							},
						}
					},
				}

				return s
			},
			expErr:   nil,
			expState: StateRefused,
		},
		{
			name: "the function returns no error and sets the state to `StateAvailable`",
			getService: func(t *testing.T) *service {
				s := newService(&ServiceConfig{
					Protocol: "http",
					Host:     "localhost",
					Port:     "8000",
				})

				s.clientPool = sync.Pool{
					New: func() any {
						return &mockHttpClient{
							mockDo: func(r *http.Request) (*http.Response, error) {
								res := &http.Response{}

								res.StatusCode = http.StatusOK

								return res, nil
							},
						}
					},
				}

				return s
			},
			expErr:   nil,
			expState: StateAvailable,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			service := tc.getService(t)

			if err := service.checkStatus(); !errors.Is(err, tc.expErr) {
				t.Errorf("expected error: %v; got error: %v\n", tc.expErr, err)
			}

			if service.state != tc.expState {
				t.Errorf("expected state: %d; got state: %v\n", tc.expState, service.state)
			}
		})
	}
}
