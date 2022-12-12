package registry

import (
	"errors"
	"testing"

	"github.com/balazskvancz/gateway/internal/service"
)

var mockServices = []service.Service{
	{
		Name:   "mock-1",
		Prefix: "/api/mock-1",
	},
	{
		Name:   "mock-2",
		Prefix: "/api/mock-2",
	},
}

const testPrefixLength = 2

var mockRegisty, _ = NewRegistry(&mockServices, 1)

func TestNewRegistry(t *testing.T) {
	tt := []struct {
		name     string
		services *[]service.Service

		expectedError error
	}{
		{
			name:          "the functions return error if the given slice is nil",
			services:      nil,
			expectedError: errServiceMapNil,
		},
		{
			name:          "the functions return error if the given slice empty",
			services:      &[]service.Service{},
			expectedError: errNoService,
		},
		{
			name:          "the functions returns the pointer of registry",
			services:      &mockServices,
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			registry, gotErr := NewRegistry(tc.services, 1)

			if tc.expectedError == nil {
				if registry == nil {
					t.Fatalf("expected not nil registry, but got one")
				}

				if gotErr != nil {
					t.Errorf("expected no error; got %v\n", gotErr)
				}

				if registry.servicePrefixL != testPrefixLength {
					t.Errorf("expected prefix length: %d; got %d\n", testPrefixLength, registry.servicePrefixL)
				}
			} else {
				if !errors.Is(gotErr, tc.expectedError) {
					t.Errorf("expected error: %v; got error: %v\n", tc.expectedError, gotErr)
				}

				if registry != nil {
					t.Errorf("expected nil registry, but got one")
				}
			}

		})
	}
}

func TestServiceByName(t *testing.T) {
	tt := []struct {
		name      string
		queryName string
		found     bool
	}{
		{
			name:      "the functions returns nil, if the requested service doesnt exist",
			queryName: "mock-3",
			found:     false,
		},
		{
			name:      "the functions returns the pointer, if the requested service exists",
			queryName: "mock-1",
			found:     true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			registry, err := NewRegistry(&mockServices, 1)

			if err != nil {
				t.Fatalf("got create error: %v\n", err)
			}

			srvc := registry.GetServiceByName(tc.queryName)

			if tc.found {
				if srvc == nil {
					t.Errorf("expected service but got nil\n")
				}
			} else {
				if srvc != nil {
					t.Errorf("expected nil but got servicservicee\n")
				}
			}
		})
	}
}

func TestAddService(t *testing.T) {
	tt := []struct {
		name     string
		registry *Registry

		toBeAdded     *service.Service
		expectedError error
	}{
		{
			name:          "the functions returns error, if the registry is nil",
			registry:      nil,
			toBeAdded:     nil,
			expectedError: errRegistryNil,
		},
		{
			name:          "the functions returns error, if the service map is nil",
			registry:      &Registry{},
			toBeAdded:     nil,
			expectedError: errServiceMapNil,
		},
		{
			name:     "the functions returns error, if the given prefix already exists",
			registry: mockRegisty,
			toBeAdded: &service.Service{
				Prefix: "/api/mock-1",
			},
			expectedError: errServiceExists,
		},
		{
			name:     "the functions doesnt return error, if it can register the service",
			registry: mockRegisty,
			toBeAdded: &service.Service{
				Prefix: "/api/mock-3",
			},
			expectedError: nil,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := tc.registry.AddService(tc.toBeAdded)

			if tc.expectedError == nil {
				if gotErr != nil {
					t.Errorf("expected no error; but got: %v\n", gotErr)
				}
			} else {
				if !errors.Is(gotErr, tc.expectedError) {
					t.Errorf("expected error: %v; got error: %v\n", tc.expectedError, gotErr)
				}

			}

		})
	}
}

func TestFindService(t *testing.T) {
	tt := []struct {
		name     string
		registry *Registry
		url      string
		found    bool
	}{
		{
			name:     "the functions returns nil, if the registry is nil",
			registry: nil,
			url:      "/api/mock-1/foo",
			found:    false,
		},
		{
			name:     "the functions returns nil, if the registry map is nil",
			registry: &Registry{},
			url:      "/api/mock-1/foo",
			found:    false,
		},
		{
			name: "the functions returns nil, if the registry has zero length",
			registry: &Registry{
				services: make(map[string]*serviceEntity),
			},
			url:   "/api/mock-1/foo",
			found: false,
		},
		{
			name:     "the functions returns nil if the prefix doesnt exists",
			registry: mockRegisty,
			url:      "/api/mock-5/foo",
			found:    false,
		},
		{
			name:     "the functions returns the pointer to service",
			registry: mockRegisty,
			url:      "/api/mock-1/foo",
			found:    true,
		},
		{
			name:     "the functions returns the pointer to service ( with missing / prefix)",
			registry: mockRegisty,
			url:      "api/mock-1/foo",
			found:    true,
		},
	}

	for _, tc := range tt {
		gotService := tc.registry.FindService(tc.url)

		if tc.found {
			if gotService == nil {
				t.Errorf("expected service; but got nil")
			}
		} else {
			if gotService != nil {
				t.Errorf("expected nil; but got service")
			}
		}
	}

}
