package gateway

import (
	"fmt"
	"time"

	"github.com/balazskvancz/rtree"
)

const (
	defaultHealthCheckFreq = time.Minute * 2
)

type registry struct {
	healthCheckFrequency time.Duration
	serviceTree          *tree
	logger
}

// Creates a new registry with empty slice of services.
func newRegistry() *registry {
	r := &registry{
		healthCheckFrequency: defaultHealthCheckFreq,
		serviceTree:          newTree(),
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

	if node := r.serviceTree.FindLongestMatch(service.Prefix); node != nil {
		return errServiceExists
	}

	return r.serviceTree.insert(service.Prefix, service)
}

// findService searches the tree based on the given url.
func (r *registry) findService(url string) *service {
	node := r.serviceTree.FindLongestMatch(url)
	if node == nil {
		return nil
	}
	return node.GetValue()
}

// getServiceByName searches for services by the given name.
func (r *registry) getServiceByName(name string) *service {
	var findServiceByName = func(n *rtree.Node[*service]) bool {
		if n == nil {
			return false
		}
		if n.GetValue() == nil {
			return false
		}
		return n.GetValue().GetValue().Name == name
	}

	node := r.serviceTree.GetByPredicate(findServiceByName)
	if node == nil {
		return nil
	}

	return node.GetValue().GetValue()
}

// Updates the status of the services, in the registry.
func (r *registry) updateStatus() {
	t := time.NewTicker(r.healthCheckFrequency)

	for {
		nodes := r.serviceTree.GetAllLeaf()

		for _, n := range nodes {
			service := n.GetValue().GetValue()

			if err := service.checkStatus(); err != nil {
				l := fmt.Sprintf("[registry] service %s – checkStatus error: %v", service.Name, err)
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
		nodes = r.serviceTree.GetAllLeaf()
		s     = make([]*service, len(nodes))
	)

	for i, node := range nodes {
		s[i] = node.GetValue().GetValue()
	}
	return s
}
