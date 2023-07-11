package gateway

import (
	"errors"
	"fmt"
	"time"
)

const (
	defaultHealthCheckFreq = time.Minute * 2
)

type registry struct {
	healthCheckFrequency time.Duration
	serviceTree          *tree
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
		healthCheckFrequency: defaultHealthCheckFreq,
		serviceTree:          newTree(),
	}
}

// addService adds the given service to the registry's tree.
func (r *registry) addService(srvc *Service) error {
	if r == nil {
		return errRegistryNil
	}

	// If the map hasnt been initialized, we return error.
	if r.serviceTree == nil {
		return errServiceTreeNil
	}

	if srvc == nil {
		return errServiceNil
	}

	if node := r.serviceTree.findLongestMatch(srvc.Prefix); node != nil {
		return errServiceExists
	}

	return r.serviceTree.insert(srvc.Prefix, srvc)
}

// findService searches the tree based on the given url.
func (r *registry) findService(url string) *Service {
	node := r.serviceTree.find(url)
	if node == nil {
		return nil
	}

	service, ok := node.value.(*Service)
	if !ok {
		return nil
	}

	return service
}

// Updates the status of the services, in the registry.
func (r *registry) updateStatus() {
	t := time.NewTicker(r.healthCheckFrequency)

	for {
		nodes := r.serviceTree.getAllLeaf()

		for _, n := range nodes {
			service, ok := n.value.(*Service)
			if !ok {
				fmt.Println("error with *Service casting")
				continue
			}

			isAvailable, err := service.CheckStatus()

			if err != nil && errors.Is(err, errServiceNotAvailable) {
				service.state = StateUnknown
				continue
			}

			if !isAvailable {
				service.state = StateRefused
				continue
			}

			service.state = StateAvailable
		}

		// Lets sleep for the given amount.
		<-t.C
	}
}
