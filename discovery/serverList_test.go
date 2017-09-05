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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandleList405 tests the list endpoint with a bad method.
func TestHandleList405(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	req, err := http.NewRequest("POST", "/list", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleList)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("expected: %v, got: %v", http.StatusMethodNotAllowed, status)
		return
	}
}

// TestHandleList401 tests the list endpoint with bad auth.
func TestHandleList401(t *testing.T) {
	auth := func(token string) bool {
		return false
	}
	server := NewServer(64646, auth)
	req, err := http.NewRequest("GET", "/list", nil)
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleList)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("expected: %v, got: %v", http.StatusUnauthorized, status)
		return
	}
}

// TestHandleList200 tests the list endpoint.
func TestHandleList200(t *testing.T) {
	server := NewServer(64646, NullAuthenticator)
	for i := 1; i <= 5; i++ {
		for j := 1; j <= 5; j++ {
			name := fmt.Sprintf("service%d", i)
			host := fmt.Sprintf("host%d", j)
			server.registry.Add(Service{Name: name, Host: host})
		}
	}
	req, err := http.NewRequest("GET", "/discover", nil)
	query := req.URL.Query()
	query.Add("name", "service1")
	req.URL.RawQuery = query.Encode()
	if err != nil {
		t.Errorf("failed to create mock request: %s", err.Error())
		return
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(server.HandleList)
	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected: %v, got: %v", http.StatusOK, status)
		return
	}
	raw, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		t.Errorf("failed to read response body: %s", err.Error())
		return
	}
	resp := struct {
		Services []Service `json:"services"`
	}{}
	err = json.Unmarshal(raw, &resp)
	if err != nil {
		t.Errorf("failed to parse json response: %s", err.Error())
		return
	}
	if length := len(resp.Services); length != 5 {
		t.Errorf("expected length: %d, got: %d\n", 5, length)
		return
	}
}
