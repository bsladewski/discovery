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
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

// handleMockSuccess mocks a simple success response.
func handleMockSuccess(w http.ResponseWriter, r *http.Request) {}

// hanldeMockError mocks a simple error response.
func handleMockError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "test error response", http.StatusInternalServerError)
}

// handleMockService mocks response that returns a service.
func handleMockService(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"service": {"name": "testName", "host": "testHost"}}`))
}

// handleMockServiceInvalid mocks response that returns a service.
func handleMockServiceInvalid(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"service": {"name": testName", "host": "testHost"}}`))
}

// handleMockServiceList mocks response that returns a service list.
func handleMockServiceList(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"services":[]}`))
}

// handleMockServiceListInvalid mocks response that returns a service list.
func handleMockServiceListInvalid(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`{"services":]}`))
}

// getMockSuccessMux gets a mux that mocks success responses.
func getMockSuccessMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleMockSuccess)
	mux.HandleFunc("/deregister", handleMockSuccess)
	mux.HandleFunc("/discover", handleMockService)
	mux.HandleFunc("/list", handleMockServiceList)
	mux.HandleFunc("/ping", handleMockSuccess)
	return mux
}

// getMockErrorMux gets a mux that mocks error responses.
func getMockErrorMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleMockError)
	mux.HandleFunc("/deregister", handleMockError)
	mux.HandleFunc("/discover", handleMockError)
	mux.HandleFunc("/list", handleMockError)
	mux.HandleFunc("/ping", handleMockSuccess)
	return mux
}

// getMockInvalidMux gets a mux that mocks invalid responses.
func getMockInvalidMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleMockError)
	mux.HandleFunc("/deregister", handleMockError)
	mux.HandleFunc("/discover", handleMockServiceInvalid)
	mux.HandleFunc("/list", handleMockServiceListInvalid)
	mux.HandleFunc("/ping", handleMockSuccess)
	return mux
}

// setupClientTest starts test servers representing different server responses.
func setupClientTest(t *testing.T) func(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	successMux := getMockSuccessMux()
	errorMux := getMockErrorMux()
	invalidMux := getMockInvalidMux()
	successMock := &http.Server{Addr: "localhost:64646", Handler: successMux}
	errorMock := &http.Server{Addr: "localhost:46464", Handler: errorMux}
	invalidMock := &http.Server{Addr: "localhost:47474", Handler: invalidMux}
	go successMock.ListenAndServe()
	go errorMock.ListenAndServe()
	go invalidMock.ListenAndServe()
	return func(t *testing.T) {
		successMock.Shutdown(context.Background())
		errorMock.Shutdown(context.Background())
		invalidMock.Shutdown(context.Background())
	}
}

// setupClientTLSTest starts a test tls server and waits for it to be ready.
func setupClientTLSTest(t *testing.T) func(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	successMux := getMockSuccessMux()
	successMock := &http.Server{Addr: "localhost:64646", Handler: successMux}
	go successMock.ListenAndServeTLS("test.crt", "test.key")
	return func(t *testing.T) {
		successMock.Shutdown(context.Background())
	}
}

// TestClientTLS tests constructing a TLS client and registry client.
func TestClientTLS(t *testing.T) {
	teardown := setupClientTLSTest(t)
	defer teardown(t)
	_, err := NewTLSClient("https://localhost:64646", "", "test.crt", false,
		5*time.Second)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	_, err = NewTLSRegistryClient("", "", "https://localhost:64646", "",
		"test.crt", false, 5*time.Second)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
}

// TestClientDiscover tests calling the discovery endpoint with a client.
func TestClientDiscover(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	table := []struct {
		target  string // target server host
		create  bool   // whether the client should be able to instantiate.
		success bool   // whether the client calls should return success.
	}{
		{"http://localhost:64646", true, true},   // success mock
		{"http://localhost:46464", true, false},  // error mock
		{"http://localhost:47474", true, false},  // invalid mock
		{"http://localhost:53535", false, false}, // invalid target
	}
	for _, row := range table {
		// construct client
		client, err := NewClient(row.target, "", time.Second)
		if err != nil && row.create {
			t.Fatalf("failed to create client: %v", err)
		} else if err == nil && !row.create {
			t.Fatalf("expected failure to create client")
		}
		// if construct client is expected to fail, continue
		if !row.create {
			continue
		}
		// test discover function
		_, err = client.Discover("")
		if err != nil && row.success {
			t.Fatalf("expected return success: %v", err)
		} else if err == nil && !row.success {
			t.Fatalf("expected return failure")
		}
	}
}

// TestClientList tests calling the list endpoint with a client.
func TestClientList(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	table := []struct {
		target  string // target server host
		create  bool   // whether the client should be able to instantiate.
		success bool   // whether the client calls should return success.
	}{
		{"http://localhost:64646", true, true},   // success mock
		{"http://localhost:46464", true, false},  // error mock
		{"http://localhost:47474", true, false},  // invalid mock
		{"http://localhost:53535", false, false}, // invalid target
	}
	for _, row := range table {
		// construct client
		client, err := NewClient(row.target, "", time.Second)
		if err != nil && row.create {
			t.Fatalf("failed to create client: %v", err)
		} else if err == nil && !row.create {
			t.Fatalf("expected failure to create client")
		}
		// if construct client is expected to fail, continue
		if !row.create {
			continue
		}
		// test list function
		_, err = client.List("")
		if err != nil && row.success {
			t.Fatalf("expected return success: %v", err)
		} else if err == nil && !row.success {
			t.Fatalf("expected return failure")
		}
	}
}

// TestClientRegister tests calling the register endpoint with a registry
// client.
func TestClientRegister(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	table := []struct {
		target  string // target server host
		create  bool   // whether the client should be able to instantiate.
		success bool   // whether the client calls should return success.
	}{
		{"http://localhost:64646", true, true},   // success mock
		{"http://localhost:46464", true, false},  // error mock
		{"http://localhost:47474", true, false},  // invalid mock
		{"http://localhost:53535", false, false}, // invalid target
	}
	for _, row := range table {
		// construct client
		client, err := NewRegistryClient("", "", row.target, "", time.Second)
		if err != nil && row.create {
			t.Fatalf("failed to create client: %v", err)
		} else if err == nil && !row.create {
			t.Fatalf("expected failure to create client")
		}
		// if construct client is expected to fail, continue
		if !row.create {
			continue
		}
		// test register function
		err = client.Register()
		if err != nil && row.success {
			t.Fatalf("expected return success: %v", err)
		} else if err == nil && !row.success {
			t.Fatalf("expected return failure")
		}
	}
}

// TestClientDeregister tests calling the deregister endpoint with a registry
// client.
func TestClientDeregister(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	table := []struct {
		target  string // target server host
		create  bool   // whether the client should be able to instantiate.
		success bool   // whether the client calls should return success.
	}{
		{"http://localhost:64646", true, true},   // success mock
		{"http://localhost:46464", true, false},  // error mock
		{"http://localhost:47474", true, false},  // invalid mock
		{"http://localhost:53535", false, false}, // invalid target
	}
	for _, row := range table {
		// construct client
		client, err := NewRegistryClient("", "", row.target, "", time.Second)
		if err != nil && row.create {
			t.Fatalf("failed to create client: %v", err)
		} else if err == nil && !row.create {
			t.Fatalf("expected failure to create client")
		}
		// if construct client is expected to fail, continue
		if !row.create {
			continue
		}
		// test deregister function
		err = client.Deregister()
		if err != nil && row.success {
			t.Fatalf("expected return success: %v", err)
		} else if err == nil && !row.success {
			t.Fatalf("expected return failure")
		}
	}
}

// TestClientAuto tests automatic registration with a registry client.
func TestClientAuto(t *testing.T) {
	teardown := setupClientTest(t)
	defer teardown(t)
	// construct client
	client, err := NewRegistryClient("", "", "http://localhost:64646", "",
		time.Second)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	// begin automatic registration
	client.Auto(10 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if !client.IsRunning() {
		t.Fatal("expected client to be running.")
	}
	// end automatic registration
	client.Deregister()
	time.Sleep(20 * time.Millisecond)
	if client.IsRunning() {
		t.Fatal("expected client to be stopped.")
	}
}
