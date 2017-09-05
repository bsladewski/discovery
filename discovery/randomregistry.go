// This is free and unencumbered software released into the public domain.

// Anyone is free to copy, modify, publish, use, compile, sell, or
// distribute this software, either in source code form or as a compiled
// binary, for any purpose, commercial or non-commercial, and by any
// means.

// In jurisdictions that recognize copyright laws, the author or authors
// of this software dedicate any and all copyright interest in the
// software to the public domain. We make this dedication for the benefit
// of the public at large and to the detriment of our heirs and
// successors. We intend this dedication to be an overt act of
// relinquishment in perpetuity of all present and future rights to this
// software under copyright law.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.

// For more information, please refer to <https://unlicense.org>

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
func (r *randomRegistry) indexOf(target Service) int {
	for i, service := range r.Services {
		if service.Name == target.Name && service.Host == target.Host {
			return i
		}
	}
	return -1
}

// getAll gets all active services of the specified name. Optionally includes
// inactive services if inactive is true.
func (r *randomRegistry) getAll(name string, inactive bool) []Service {
	var (
		services []Service
		stale    []Service
	)
	r.mutex.Lock()
	for _, service := range r.Services {
		if name == "" || name == service.Name {
			if time.Since(service.Added) < r.Timeout ||
				(inactive && time.Since(service.Added) >= r.Timeout &&
					time.Since(service.Added) < r.Keep) {
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

func (r *randomRegistry) Add(service Service) {
	r.mutex.Lock()
	if idx := r.indexOf(service); idx >= 0 {
		r.Services[idx].Added = time.Now()
	} else {
		service.Added = time.Now()
		r.Services = append(r.Services, service)
	}
	r.mutex.Unlock()
}

func (r *randomRegistry) Remove(service Service) {
	r.mutex.Lock()
	if idx := r.indexOf(service); idx >= 0 {
		r.Services = append(r.Services[:idx], r.Services[idx+1:]...)
	}
	r.mutex.Unlock()
}

func (r *randomRegistry) Get(name string) (Service, error) {
	services := r.getAll(name, false)
	if len(services) == 0 {
		return Service{}, fmt.Errorf("so such service '%s'", name)
	}
	return services[rand.Intn(len(services))], nil
}

func (r *randomRegistry) List(name string) []Service {
	return r.getAll(name, true)
}

func (r *randomRegistry) SetTimeout(timeout time.Duration) {
	r.mutex.Lock()
	r.Timeout = timeout
	r.mutex.Unlock()
}

func (r *randomRegistry) SetKeep(keep time.Duration) {
	r.mutex.Lock()
	r.Keep = keep
	r.mutex.Unlock()
}

// NewRandomRegistry creates a Registry that load balances by selecting a
// random service when replicants exist.
func NewRandomRegistry(timeout time.Duration, keep time.Duration) Registry {
	return &randomRegistry{
		Services: make([]Service, 0),
		Timeout:  timeout,
		Keep:     keep,
		mutex:    &sync.Mutex{},
	}
}
