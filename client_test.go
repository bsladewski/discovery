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
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"testing"
	"time"
)

// setupClientTest starts a test server and waits for it to be ready.
func setupClientTest(t *testing.T) func(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	server := NewRandomServer(64646, NullAuthenticator)
	server.registry.Add(Service{Name: "service", Host: "host"})
	go server.ListenAndServe()
	timestamp := time.Now()
	for time.Since(timestamp) > 5*time.Second {
		_, err := net.DialTimeout("tcp", "http://localhost:64646", 10*time.Millisecond)
		if err == nil {
			break
		} else {
			fmt.Println(err.Error())
		}
	}
	return func(t *testing.T) {
		server.Shutdown(context.Background())
	}
}

// TestClientDiscover tests calling the discovery endpoint with a client.
func TestClientDiscover(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	client, err := NewClient("http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	_, err = client.Discover("service")
	if err != nil {
		t.Errorf("failed to get service: %s", err.Error())
		return
	}
	_, err = client.Discover("invalid")
	if err == nil {
		t.Errorf("failed to get service: %s", err.Error())
		return
	}
}

// TestClientDiscover tests calling the list endpoint with a client.
func TestClientList(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	client, err := NewClient("http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	_, err = client.List("service")
	if err != nil {
		t.Errorf("failed to get service: %s", err.Error())
		return
	}
}

// TestClientRegister tests calling the register endpoint with a registry
// client.
func TestClientRegister(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	client, err := NewRegistryClient("service", "hostName",
		"http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	err = client.Register()
	if err != nil {
		t.Errorf("failed to register service: %s", err.Error())
		return
	}
}

// TestClientDeregister tests calling the deregister endpoint with a registry
// client.
func TestClientDeregister(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	client, err := NewRegistryClient("service", "hostName",
		"http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	err = client.Deregister()
	if err != nil {
		t.Errorf("failed to register service: %s", err.Error())
		return
	}
}

// TestClientAuto tests automatic registration with a registry client.
func TestClientAuto(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	client, err := NewRegistryClient("service", "hostName",
		"http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	client.Auto(10 * time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	err = client.Deregister()
	if err != nil {
		t.Errorf("failed to register service: %s", err.Error())
		return
	}
	time.Sleep(50 * time.Millisecond)
	if client.IsRunning() {
		t.Errorf("client still running")
		return
	}
}
