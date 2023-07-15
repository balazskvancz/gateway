package gateway

import (
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
func newRegistry(opts ...registryOptionFunc) *registry {
	r := &registry{
		healthCheckFrequency: defaultHealthCheckFreq,
		serviceTree:          newTree(),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

// addService adds the given service to the registry's tree.
func (r *registry) addService(srvc *service) error {
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
func (r *registry) findService(url string) *service {
	node := r.serviceTree.find(url)
	if node == nil {
		return nil
	}

	service, ok := node.value.(*service)
	if !ok {
		return nil
	}

	return service
}

// getServiceByName searches for services by the given name.
func (r *registry) getServiceByName(name string) *service {
	node := r.serviceTree.getByPredicate(func(n *node) bool {
		s, ok := n.value.(*service)
		if !ok {
			return false
		}
		return s.Name == name
	})

	if node == nil {
		return nil
	}

	serv, ok := node.value.(*service)
	if !ok {
		return nil
	}

	return serv
}

// Updates the status of the services, in the registry.
func (r *registry) updateStatus() {
	t := time.NewTicker(r.healthCheckFrequency)

	for {
		nodes := r.serviceTree.getAllLeaf()

		for _, n := range nodes {
			service, ok := n.value.(*service)
			if !ok {
				fmt.Println("error with *Service casting")
				continue
			}

			if err := service.checkStatus(); err != nil {
				fmt.Printf("[registry] service %s – checkStatus error: %v\n", service.Name, err)
			}
		}

		// Lets sleep for the given amount.
		<-t.C
	}
}

// setServiceAvailable changes the state of service matched by
// given name to StateAvailable.
func (r *registry) setServiceAvailable(name string) {
	service := r.getServiceByName(name)
	if service == nil {
		// No match, no effect.
		return
	}

	service.setState(StateAvailable)
}
