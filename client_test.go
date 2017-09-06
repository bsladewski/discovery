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
	"testing"
	"time"
)

// TestClientDiscover tests calling the discovery endpoint with a client.
func TestClientDiscover(t *testing.T) {
	server := NewRandomServer(64646, NullAuthenticator)
	server.registry.Add(Service{Name: "service", Host: "hostName"})
	go server.Run()
	ctx := context.Background()
	defer server.Shutdown(ctx)
	client, err := NewClient("http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	host, err := client.Discover("service")
	if err != nil {
		t.Errorf("failed to get service: %s", err.Error())
		return
	}
	if host != "hostName" {
		t.Errorf("expected: hostName, got: %s", host)
		return
	}
}

// TestClientDiscover tests calling the list endpoint with a client.
func TestClientList(t *testing.T) {
	server := NewRandomServer(64646, NullAuthenticator)
	for i := 1; i <= 5; i++ {
		server.registry.Add(Service{
			Name: "service",
			Host: fmt.Sprintf("hostName%d", i),
		})
	}
	go server.Run()
	ctx := context.Background()
	defer server.Shutdown(ctx)
	client, err := NewClient("http://localhost:64646", "", 10*time.Second)
	if err != nil {
		t.Errorf("failed to create client: %s", err.Error())
		return
	}
	services, err := client.List("service")
	if err != nil {
		t.Errorf("failed to get service: %s", err.Error())
		return
	}
	if length := len(services); length != 5 {
		t.Errorf("expected: 5, got: %d", length)
		return
	}
}
