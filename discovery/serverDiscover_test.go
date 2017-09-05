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
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleDiscover405 tests the discover endpoint with a bad method.
func TestHandleDiscover405(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("POST", "/discover", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleDiscover)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected: %v, got: %v", http.StatusMethodNotAllowed, status)
		return
	}
}

// TestHandleDiscover401 tests the discover endpoint with bad auth.
func TestHandleDiscover401(t *testing.T) {
	auth := func(token string) bool {
		return false
	}
	server := NewServer(64646, auth)
	req, err := http.NewRequest("GET", "/discover", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleDiscover)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected: %v, got: %v", http.StatusUnauthorized, status)
		return
	}
}

// TestHandleDiscover400 tests the discover endpoint with a bad request.
func TestHandleDiscover400(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("GET", "/discover", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleDiscover)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected: %v, got: %v", http.StatusBadRequest, status)
		return
	}
}

// TestHandleDiscover404 tests the discover endpoint with a nonexistant service.
func TestHandleDiscover404(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("GET", "/discover", nil)
	query := req.URL.Query()
	query.Add("name", "nonexistant")
	req.URL.RawQuery = query.Encode()
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleDiscover)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("expected: %v, got: %v", http.StatusNotFound, status)
		return
	}
}

// TestHandleDiscover200 tests the discover endpoint.
func TestHandleDiscover200(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	server.registry.Add(Service{Name: "test_service", Host: "localhost:78781"})
	req, err := http.NewRequest("GET", "/discover", nil)
	query := req.URL.Query()
	query.Add("name", "test_service")
	req.URL.RawQuery = query.Encode()
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleDiscover)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected: %v, got: %v", http.StatusOK, status)
		return
	}
}
