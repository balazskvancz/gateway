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
	serviceTree          *tree[*service]
	logger
}

// Creates a new registry with empty slice of services.
func newRegistry() *registry {
	r := &registry{
		healthCheckFrequency: defaultHealthCheckFreq,
		serviceTree:          newTree[*service](),
	}

	return r
}

func (r *registry) withHealthCheck(freq time.Duration) {
	r.healthCheckFrequency = freq
}

func (r *registry) withLogger(l logger) {
	r.logger = l
}

// addService adds the given service to the registry's tree.
func (r *registry) addService(conf *ServiceConfig) error {
	if err := validateService(conf); err != nil {
		return err
	}

	if r == nil {
		return errRegistryNil
	}

	// If the map hasnt been initialized, we return error.
	if r.serviceTree == nil {
		return errServiceTreeNil
	}

	service := newService(conf)

	if node := r.serviceTree.findLongestMatch(service.Prefix); node != nil {
		return errServiceExists
	}

	return r.serviceTree.insert(service.Prefix, service)
}

// findService searches the tree based on the given url.
func (r *registry) findService(url string) *service {
	node := r.serviceTree.findLongestMatch(url)
	if node == nil {
		return nil
	}
	return node.value
}

// getServiceByName searches for services by the given name.
func (r *registry) getServiceByName(name string) *service {
	var findServiceByName = func(n *node[*service]) bool {
		if n == nil {
			return false
		}
		if n.value == nil {
			return false
		}
		return n.value.Name == name
	}

	node := r.serviceTree.getByPredicate(findServiceByName)
	if node == nil {
		return nil
	}

	return node.value
}

// Updates the status of the services, in the registry.
func (r *registry) updateStatus() {
	t := time.NewTicker(r.healthCheckFrequency)

	for {
		nodes := r.serviceTree.getAllLeaf()

		for _, n := range nodes {
			service := n.value

			if err := service.checkStatus(); err != nil {
				l := fmt.Sprintf("[registry] service %s – checkStatus error: %v\n", service.Name, err)
				r.logger.Error(l)
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

func (r *registry) getAllServices() []*service {
	var (
		nodes = r.serviceTree.getAllLeaf()
		s     = make([]*service, len(nodes))
	)

	for i, node := range nodes {
		s[i] = node.value
	}
	return s
}
