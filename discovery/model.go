// Package discovery implements a service registry for tracking the location of
// distributed microservices.
package discovery

import "time"

// Service holds information about a service as well as the last time the
// service was renewed.
type Service struct {
	Name  string    `json:"name"`
	Host  string    `json:"host"`
	Added time.Time `json:"added"`
}

// Registry holds host names for services by name.
type Registry interface {
	Add(service Service)              // Add adds or updates a service to this registry.
	Remove(service Service)           // Remove removes a service from this registry.
	Get(name string) (Service, error) // Get gets the specified service.
	List(name string) []Service       // List gets all services filtered by name.
	SetTimeout(timeout time.Duration) // SetTimeout updates the timeout duration.
	SetKeep(timeout time.Duration)    // SetKeep updates the keep duration.
}

// Authenticator defines how to handle http authentication.
type Authenticator func(token string) bool

// NullAuthenticator the authenticator that always returns true.
func NullAuthenticator(token string) bool { return true }
