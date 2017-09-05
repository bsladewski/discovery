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
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleRegister405 tests the register endpoint with a bad method.
func TestHandleRegister405(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("GET", "/register", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleRegister)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected: %v, got: %v", http.StatusMethodNotAllowed, status)
		return
	}
}

// TestHandleRegister401 tests the register endpoint with bad auth.
func TestHandleRegister401(t *testing.T) {
	auth := func(token string) bool {
		return false
	}
	server := NewServer(64646, auth)
	req, err := http.NewRequest("POST", "/register", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleRegister)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected: %v, got: %v", http.StatusUnauthorized, status)
		return
	}
}

// TestHandleRegister400 tests the register endpoint with a bad request.
func TestHandleRegister400(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("POST", "/register", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleRegister)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("expected: %v, got: %v", http.StatusBadRequest, status)
		return
	}
}

// TestHandleRegister200 tests the register endpoint.
func TestHandleRegister200(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	service := Service{Name: "service1", Host: "host1"}
	raw, err := json.Marshal(service)
	if err != nil {
		t.Errorf("failed to create request body: %s", err.Error())
		return
	}
	req, err := http.NewRequest("POST", "/register", bytes.NewBuffer(raw))
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleRegister)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected: %v, got: %v", http.StatusOK, status)
		return
	}
	if _, err := server.registry.Get(service.Name); err != nil {
		t.Errorf("failed to retrieve registered service: %s", err.Error())
		return
	}
}
