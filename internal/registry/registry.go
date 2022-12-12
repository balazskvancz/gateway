package registry

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/balazskvancz/gateway/internal/service"
	"github.com/balazskvancz/gateway/internal/utils"
)

var (
	errNoService     = errors.New("zero length of services")
	errRegistryNil   = errors.New("registry is nil")
	errServiceExists = errors.New("service already registered")
	errServiceMapNil = errors.New("service registry is nil")
	errServiceNil    = errors.New("service is nil")
)

const (
	StateRegistered = iota
	StateUnknown
	StateRefused
	StateAvailable
)

type serviceEntity struct {
	service *service.Service

	state uint8
}

type Registry struct {
	services map[string]*serviceEntity
	sleepMin uint8

	servicePrefixL int8
}

// Creates a new registry with empty slice of services.
func NewRegistry(services *[]service.Service, sleepMin uint8) (*Registry, error) {
	if services == nil {
		return nil, errServiceMapNil
	}

	if len(*services) == 0 {
		return nil, errNoService
	}

	servicesMap := make(map[string]*serviceEntity)

	for _, serv := range *services {
		if _, exists := servicesMap[serv.Prefix]; exists {
			fmt.Printf("Duplicate service ([%s] => %s). Ignoring.\n", serv.Name, serv.Prefix)

			continue
		}

		serv := &service.Service{
			Protocol:     serv.Protocol,
			Name:         serv.Name,
			Host:         serv.Host,
			Port:         serv.Port,
			Prefix:       serv.Prefix,
			CContentType: serv.CContentType,
		}

		// Create the associating http client.
		serv.CreateClient()

		servicesMap[serv.Prefix] = &serviceEntity{
			service: serv,
			state:   StateRegistered, /*StateAvailable */
		}
	}

	var length int8 = 0

	for key := range servicesMap {
		splitted := utils.GetUrlParts(key)

		length = int8(len(splitted))

		break
	}

	return &Registry{
		services:       servicesMap,
		servicePrefixL: length,
		sleepMin:       sleepMin,
	}, nil
}

// Adds a new service to the registry.
func (r *Registry) AddService(srvc *service.Service) error {
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
	_, exists := r.services[srvc.Prefix]

	if exists {
		return errServiceExists
	}

	r.services[srvc.Prefix] = &serviceEntity{
		service: srvc,
		state:   StateRegistered,
	}

	return nil
}

func (r *Registry) findService(prefix string) *serviceEntity {
	sEntity, exists := r.services[prefix]

	if !exists {
		return nil
	}

	return sEntity
}

// Finds the service based on url.
func (r *Registry) FindService(url string) *serviceEntity {
	// Simple checking for not calling anything on a nil pointer.
	if r == nil {
		return nil
	}

	if r.services == nil {
		return nil
	}

	if len(r.services) == 0 {
		return nil
	}

	// If the given url doesnt start with "/", put it there.
	if !strings.HasPrefix(url, "/") {
		url = "/" + url
	}

	splitted := utils.GetUrlParts(url)
	if int8(len(splitted)) < r.servicePrefixL {
		return nil
	}

	// Creates a prefix which is suitable	for the stored services.
	sPrefix := "/" + strings.Join(splitted[:r.servicePrefixL], "/")

	return r.findService(sPrefix)
}

// Finds and returns a service by name.
func (r *Registry) GetServiceByName(name string) *serviceEntity {
	for _, v := range r.services {
		if v.service.Name == name {
			return v
		}
	}

	return nil
}

// Updates the status of the services, in the registry.
func (r *Registry) UpdateStatus() {
	for {
		for _, s := range r.services {
			isAvailable, err := s.GetService().CheckStatus()

			if err != nil && errors.Is(err, service.ErrServiceNotAvailable) {
				r.setState(s.GetService().Prefix, StateUnknown)
				continue
			}

			if !isAvailable {
				r.setState(s.GetService().Prefix, StateRefused)
				continue
			}

			r.setState(s.GetService().Prefix, StateAvailable)
		}

		sleepTime := time.Duration(r.sleepMin) * time.Minute
		time.Sleep(sleepTime)
	}
}

func (r *Registry) setState(serv string, state uint8) {
	service, exits := r.services[serv]

	if !exits {
		return
	}

	r.services[serv] = &serviceEntity{
		service: service.service,
		state:   state,
	}
}

// ------------------

// Returns the service itself from the entity.
func (sE *serviceEntity) GetService() *service.Service {
	return sE.service
}

// Returns whether the service is available inside the entity.
func (sE *serviceEntity) IsAvailable() bool {
	return sE.state == StateAvailable
}

// Returns the state of service inside the entity.
func (sE *serviceEntity) GetState() uint8 {
	return sE.state
}
