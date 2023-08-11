package gateway

import (
	"errors"
	"testing"
)

func TestAddService(t *testing.T) {
	type (
		registryFactory func(*testing.T) *registry
	)

	type testCase struct {
		name        string
		getRegistry registryFactory
		conf        *ServiceConfig
		expError    error
	}

	var goodConfig = &ServiceConfig{
		Protocol: "http",
		Name:     "mockService",
		Host:     "localhost",
		Port:     "3000",
		Prefix:   "/mock",
	}

	tt := []testCase{
		{
			name:        "the function returns error if the serviceConfig is nil",
			getRegistry: func(t *testing.T) *registry { return nil },
			conf:        nil,
			expError:    errConfigIsNil,
		},
		{
			name:        "the function returns error if the the registry is nil",
			getRegistry: func(t *testing.T) *registry { return nil },
			conf:        goodConfig,
			expError:    errRegistryNil,
		},
		{
			name: "the function returns error if tree of the registry is nil",
			getRegistry: func(t *testing.T) *registry {
				return &registry{}
			},
			conf:     goodConfig,
			expError: errServiceTreeNil,
		},
		{
			name: "the function returns error if the given service is already contained",
			getRegistry: func(t *testing.T) *registry {
				r := newRegistry()

				if err := r.addService(goodConfig); err != nil {
					t.Fatalf("cant init service, err: %v\n", err)
				}

				return r
			},
			conf:     goodConfig,
			expError: errServiceExists,
		},
		{
			name: "the function does not return error",
			getRegistry: func(t *testing.T) *registry {
				return newRegistry()
			},
			conf:     goodConfig,
			expError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				reg = tc.getRegistry(t)
			)

			if err := reg.addService(tc.conf); !errors.Is(err, tc.expError) {
				t.Errorf("expected error: %v; got: %v\n", tc.expError, err)
			}
		})
	}
}

func TestFindService(t *testing.T) {
	type (
		registryFactory func(*testing.T) *registry
	)

	type testCase struct {
		name        string
		getRegistry registryFactory
		isMatch     bool
	}

	tt := []testCase{
		{
			name:        "the function returns nil in an empty registry",
			getRegistry: func(t *testing.T) *registry { return newRegistry() },
			isMatch:     false,
		},
		{
			name: "the function returns nil in a non empty registry",
			getRegistry: func(t *testing.T) *registry {
				r := newRegistry()

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-1",
					Host:     "localhost",
					Port:     "3000",
					Prefix:   "/foo/baz",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				return r
			},
			isMatch: false,
		},
		{
			name: "the function returns a service in a non empty registry",
			getRegistry: func(t *testing.T) *registry {
				r := newRegistry()

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-1",
					Host:     "localhost",
					Port:     "3000",
					Prefix:   "/foo/bar",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-2",
					Host:     "localhost",
					Port:     "3010",
					Prefix:   "/foo/baz",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				return r
			},
			isMatch: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				reg = tc.getRegistry(t)
			)

			serv := reg.findService("/foo/bar")

			if tc.isMatch && serv == nil {
				t.Error("expected to find, but did not")
			}

			if !tc.isMatch && serv != nil {
				t.Error("not expected to find, but did")
			}
		})
	}
}

func TestGetServiceByName(t *testing.T) {
	type (
		registryFactory func(*testing.T) *registry
	)

	type testCase struct {
		name        string
		getRegistry registryFactory
		isMatch     bool
	}

	tt := []testCase{
		{
			name:        "the function returns nil in an empty registry",
			getRegistry: func(t *testing.T) *registry { return newRegistry() },
			isMatch:     false,
		},
		{
			name: "the function returns nil in a non empty registry",
			getRegistry: func(t *testing.T) *registry {
				r := newRegistry()

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-1",
					Host:     "localhost",
					Port:     "3000",
					Prefix:   "/foo/baz",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				return r
			},
			isMatch: false,
		},
		{
			name: "the function returns a service in a non empty registry",
			getRegistry: func(t *testing.T) *registry {
				r := newRegistry()

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-1",
					Host:     "localhost",
					Port:     "3000",
					Prefix:   "/foo/bar",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				if err := r.addService(&ServiceConfig{
					Protocol: "http",
					Name:     "mock-name-2",
					Host:     "localhost",
					Port:     "3010",
					Prefix:   "/foo/baz",
				}); err != nil {
					t.Fatalf("expected not to get error; but got: %v\n", err)
				}

				return r
			},
			isMatch: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var (
				reg = tc.getRegistry(t)
			)

			serv := reg.getServiceByName("mock-name-2")

			if tc.isMatch && serv == nil {
				t.Error("expected to find, but did not")
			}

			if !tc.isMatch && serv != nil {
				t.Error("not expected to find, but did")
			}
		})
	}
}
