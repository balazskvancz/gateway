package gateway

import (
	"errors"
	"time"
)

const (
	defaultHealthCheckFreq = time.Minute * 2
)

type registry struct {
	healthCheckFrequency time.Duration
	services             map[string]*Service

	serviceTree *tree
}

type registryOptionFunc func(*registry)

func withHealthCheck(freq time.Duration) registryOptionFunc {
	return func(r *registry) {
		r.healthCheckFrequency = freq
	}
}

// Creates a new registry with empty slice of services.
func newRegistry() *registry {
	return &registry{
		services:             make(map[string]*Service),
		healthCheckFrequency: defaultHealthCheckFreq,
	}
}

// Adds a new service to the registry.
func (r *registry) addService(srvc *Service) error {
	if r == nil {
		return errRegistryNil
	}

	// If the map hasnt been initialized, we return error.
	if r.services == nil {
		return errServiceMapNil
	}

	if srvc == nil {
		return errServiceNil
	}

	// Check if already registered.
	if _, exists := r.services[srvc.Prefix]; exists {
		return errServiceExists
	}

	r.services[srvc.Prefix] = srvc

	return nil
}

func (r *registry) findServiceByPrefix(prefix string) *Service {
	service, exists := r.services[prefix]
	if !exists {
		return nil
	}

	return service
}

// Finds the service based on url.
func (r *registry) FindService(url string) *Service {
	return nil
}

// getServiceByName finds and returns a service by name.
func (r *registry) getServiceByName(name string) *Service {
	for _, s := range r.services {
		if s.Name == name {
			return s
		}
	}

	return nil
}

// Updates the status of the services, in the registry.
func (r *registry) updateStatus() {
	t := time.Tick(r.healthCheckFrequency)

	for {
		for _, s := range r.services {
			isAvailable, err := s.CheckStatus()

			if err != nil && errors.Is(err, errServiceNotAvailable) {
				r.setState(s.Prefix, StateUnknown)
				continue
			}

			if !isAvailable {
				r.setState(s.Prefix, StateRefused)
				continue
			}

			r.setState(s.Prefix, StateAvailable)
		}

		// Lets sleep for the given amount.
		<-t
	}
}

func (r *registry) setState(serv string, state serviceState) {
	_, exits := r.services[serv]

	if !exits {
		return
	}

	// r.services[serv] =
}
