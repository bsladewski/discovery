// Package discovery implements a service registry for tracking the location of
// distributed microservices.
package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Server represents an http interface to a service registry.
type Server struct {
	registry      Registry
	port          int
	authenticator Authenticator

	tls      bool
	certFile string
	keyFile  string
}

// HandleRegister adds a service to or renews a service with the registry.
func (server *Server) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	if !server.authenticator(r.Header.Get("Authentication")) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	service := Service{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&service)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	server.registry.Add(service)
}

// HandleDeregister removes a service from the registry.
func (server *Server) HandleDeregister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	if !server.authenticator(r.Header.Get("Authentication")) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	service := Service{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&service)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	server.registry.Remove(service)
}

// HandleDiscover gets a service from the registry.
func (server *Server) HandleDiscover(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	if !server.authenticator(r.Header.Get("Authentication")) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	name := r.URL.Query().Get("name")
	if name == "" {
		http.Error(w, "no service name provided", http.StatusBadRequest)
		return
	}
	service, err := server.registry.Get(name)
	if err != nil {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}
	raw, err := json.Marshal(service)
	if err != nil {
		http.Error(w, "failed to write service", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// HandleList lists all services registered with the registry.
func (server *Server) HandleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "method not supported", http.StatusMethodNotAllowed)
		return
	}
	if !server.authenticator(r.Header.Get("Authorization")) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	resp := struct {
		Services []Service `json:"services"`
	}{}
	resp.Services = server.registry.List(r.URL.Query().Get("name"))
	raw, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to write services", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

// Run registers the http endpoints and runs the servers. Returns error on exit.
func (server *Server) Run() error {
	http.HandleFunc("/register", server.HandleRegister)
	http.HandleFunc("/deregister", server.HandleDeregister)
	http.HandleFunc("/discover", server.HandleDiscover)
	http.HandleFunc("/list", server.HandleList)
	addr := fmt.Sprintf("localhost:%d", server.port)
	if server.tls {
		return http.ListenAndServeTLS(addr, server.certFile, server.keyFile, nil)
	}
	return http.ListenAndServe(addr, nil)
}

// SetTimeout updates how long a service should be considered active.
func (server *Server) SetTimeout(timeout time.Duration) {
	server.registry.SetTimeout(timeout)
}

// SetKeep updates how long the registry should keep inactive services.
func (server *Server) SetKeep(keep time.Duration) {
	server.registry.SetKeep(keep)
}

// NewServer returns a server on the specified port. Takes an authenticator that
// defines how authentication is handled.
func NewServer(port int, authenticator Authenticator) *Server {
	return &Server{
		registry:      NewRandomRegistry(30*time.Minute, 24*time.Hour),
		port:          port,
		authenticator: authenticator,
	}
}

// NewTLSServer returns an encrypted server on the specified port. Takes and
// authenticator that defines how authentication is handled as well as the paths
// to a certificate and key file.
func NewTLSServer(port int, authenticator Authenticator, certFile,
	keyFile string) *Server {
	return &Server{
		registry:      NewRandomRegistry(30*time.Minute, 24*time.Hour),
		port:          port,
		authenticator: authenticator,
		tls:           true,
		certFile:      certFile,
		keyFile:       keyFile,
	}
}
