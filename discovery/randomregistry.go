// Package discovery implements a service registry for tracking the location of
// distributed microservices.
package discovery

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// randomRegistry implements Registry with a random load balancing algorithm.
type randomRegistry struct {
	Services []Service
	Timeout  time.Duration
	Keep     time.Duration
	mutex    *sync.Mutex
}

// indexOf gets the index of the specified service in the registry or -1.
func (r randomRegistry) indexOf(target Service) int {
	for i, service := range r.Services {
		if service.Name == target.Name {
			return i
		}
	}
	return -1
}

func (r randomRegistry) Add(service Service) {
	r.mutex.Lock()
	if idx := r.indexOf(service); idx >= 0 {
		r.Services[idx].Added = time.Now()
	} else {
		r.Services = append(r.Services, service)
	}
	r.mutex.Unlock()
}

func (r randomRegistry) Remove(service Service) {
	r.mutex.Lock()
	if idx := r.indexOf(service); idx >= 0 {
		r.Services = append(r.Services[:idx], r.Services[idx+1:]...)
	}
	r.mutex.Unlock()
}

// getAll gets all active services of the specified name. Gets active services
// if active is true, inactive services if active is false.
func (r randomRegistry) getAll(name string, active bool) []Service {
	var (
		services []Service
		stale    []Service
	)
	r.mutex.Lock()
	for _, service := range r.Services {
		if service.Name == name {
			if (active && time.Since(service.Added) < r.Timeout) ||
				(!active && time.Since(service.Added) >= r.Timeout) {
				services = append(services, service)
			}
		}
		if time.Since(service.Added) > r.Keep {
			stale = append(stale, service)
		}
	}
	r.mutex.Unlock()
	for _, service := range stale {
		r.Remove(service)
	}
	return services
}

func (r randomRegistry) Get(name string) (Service, error) {
	services := r.getAll(name, true)
	if len(services) == 0 {
		return Service{}, fmt.Errorf("so such service '%s'", name)
	}
	return services[rand.Intn(len(services))], nil
}

func (r randomRegistry) List(name string) []Service {
	if name == "" {
		return append([]Service{}, r.Services...)
	}
	return append(r.getAll(name, true), r.getAll(name, false)...)
}

func (r randomRegistry) SetTimeout(timeout time.Duration) {
	r.mutex.Lock()
	r.Timeout = timeout
	r.mutex.Unlock()
}

func (r randomRegistry) SetKeep(keep time.Duration) {
	r.mutex.Lock()
	r.Keep = keep
	r.mutex.Unlock()
}

// NewRandomRegistry creates a Registry that load balances by selecting a
// random service when replicants exist.
func NewRandomRegistry(timeout time.Duration, keep time.Duration) Registry {
	return randomRegistry{
		Services: make([]Service, 0),
		Timeout:  timeout,
		Keep:     keep,
		mutex:    &sync.Mutex{},
	}
}
