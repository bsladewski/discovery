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
	"testing"
	"time"
)

// generateTestRegistry generates a random registry with the specified services
// and replicants for each service.
func generateTestRegistry(serviceCount, replicantCount int) randomRegistry {
	registry := NewRandomRegistry(12*time.Hour, 24*time.Hour).(*randomRegistry)
	services := []Service{}
	for i := 1; i <= serviceCount; i++ {
		for j := 1; j <= replicantCount; j++ {
			name := fmt.Sprintf("service%d", i)
			host := fmt.Sprintf("host%d", j)
			service := Service{Name: name, Host: host, Added: time.Now()}
			services = append(services, service)
		}
	}
	registry.Services = services
	return *registry
}

// TestIndexOf tests the randomRegistry.indexOf function.
func TestIndexOf(t *testing.T) {
	registry := generateTestRegistry(5, 2)
	table := []struct {
		service  Service
		expected int
	}{
		{service: Service{Name: "service3", Host: "host2"}, expected: 5},
		{service: Service{Name: "serviceX", Host: "hostX"}, expected: -1},
	}
	for _, row := range table {
		if idx := registry.indexOf(row.service); idx != row.expected {
			t.Fatalf("expected index: %d, got: %d; %v", row.expected, idx,
				row)
		}
	}
}

// TestGetAll tests the randomRegistry.getAll function.
func TestGetAll(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	for i, service := range registry.Services {
		if i%5 == 0 {
			service.Added = service.Added.Add(-13 * time.Hour)
			registry.Services[i] = service
		}
	}
	table := []struct {
		name        string
		inactive    bool
		expectedLen int
	}{
		{name: "service1", expectedLen: 4},
		{name: "service1", inactive: true, expectedLen: 5},
		{name: "invalid", expectedLen: 0},
		{name: "", expectedLen: 20},
		{name: "", inactive: true, expectedLen: 25},
	}
	for _, row := range table {
		services := registry.getAll(row.name, row.inactive)
		if length := len(services); length != row.expectedLen {
			t.Fatalf("expected: %d, got: %d; %v", row.expectedLen, length,
				row)
		}
	}
}

// TestStale test removing stale entries on randomRegistry.getAll call.
func TestStale(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	for i, service := range registry.Services {
		service.Added = service.Added.Add(-25 * time.Hour)
		registry.Services[i] = service
	}
	services := registry.getAll("", true)
	if length := len(services); length > 0 {
		t.Fatalf("expected empty list, got length: %d", length)
	}
	if length := len(registry.Services); length > 0 {
		t.Fatalf("expected empty registry, got length: %d", length)
	}
}

// TestAdd tests the randomRegistry.Add function.
func TestAdd(t *testing.T) {
	registry := generateTestRegistry(0, 0)
	table := []struct {
		service     Service
		expectedLen int
		expectedIdx int
	}{
		{service: Service{Name: "service1", Host: "host1"}, expectedLen: 1,
			expectedIdx: 0},
		{service: Service{Name: "service1", Host: "host2"}, expectedLen: 2,
			expectedIdx: 1},
		{service: Service{Name: "service1", Host: "host1"}, expectedLen: 2,
			expectedIdx: 0},
		{service: Service{Name: "service2", Host: "host1"}, expectedLen: 3,
			expectedIdx: 2},
	}
	for _, row := range table {
		registry.Add(row.service)
		if length := len(registry.Services); length != row.expectedLen {
			t.Fatalf("expected length: %d, got: %d; %v", row.expectedLen,
				length, row)
		}
		if idx := registry.indexOf(row.service); idx != row.expectedIdx {
			t.Fatalf("expected index: %d, got: %d; %v", row.expectedIdx, idx,
				row)
		}
	}
}

// TestRemove tests the randomRegistry.Remove function.
func TestRemove(t *testing.T) {
	registry := generateTestRegistry(1, 5)
	table := []struct {
		service     Service
		expectedLen int
	}{
		{service: Service{Name: "invalid", Host: "invalid"}, expectedLen: 5},
		{service: Service{Name: "service1", Host: "host3"}, expectedLen: 4},
		{service: Service{Name: "service1", Host: "host3"}, expectedLen: 4},
	}
	for _, row := range table {
		registry.Remove(row.service)
		if length := len(registry.Services); length != row.expectedLen {
			t.Fatalf("expected length: %d, got: %d; %v", row.expectedLen,
				length, row)
		}
		if idx := registry.indexOf(row.service); idx != -1 {
			t.Fatalf("expected index: -1, got: %d; %v", idx, row)
		}
	}
}

// TestGet tests the randomRegistry.Get function.
func TestGet(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	table := []struct {
		name        string
		expectedErr bool
	}{
		{name: "service3", expectedErr: false},
		{name: "invalid", expectedErr: true},
	}
	for _, row := range table {
		service, err := registry.Get(row.name)
		if !row.expectedErr && service.Name != row.name {
			t.Fatalf("expected: %s, got: %s; %v", row.name, service.Name, row)
		}
		if row.expectedErr && err == nil {
			t.Fatalf("expected error, got nil; %v", row)
		}
	}
}

// TestList tests the randomRegistry.List function.
func TestList(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	for i, service := range registry.Services {
		if i%5 == 0 {
			service.Added = service.Added.Add(-13 * time.Hour)
			registry.Services[i] = service
		}
	}
	table := []struct {
		name        string
		expectedLen int
	}{
		{name: "service1", expectedLen: 5},
		{name: "invalid", expectedLen: 0},
		{name: "", expectedLen: 25},
	}
	for _, row := range table {
		services := registry.List(row.name)
		if length := len(services); length != row.expectedLen {
			t.Fatalf("expected: %d, got: %d; %v", row.expectedLen, length,
				row)
		}
	}
}

// TestSetTimeout tests the randomRegistry.SetTimeout function.
func TestSetTimeout(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	for i, service := range registry.Services {
		service.Added = service.Added.Add(-7 * time.Hour)
		registry.Services[i] = service
	}
	services := registry.getAll("", false)
	if length := len(services); length != 25 {
		t.Fatalf("failed to propagate registry, got length: %d", length)
	}
	registry.SetTimeout(6 * time.Hour)
	services = registry.getAll("", false)
	if length := len(services); length != 0 {
		t.Fatalf("expected empty list, got length: %d\n", length)
	}
}

// TestSetKeep tests the randomRegistry.SetKeep function.
func TestSetKeep(t *testing.T) {
	registry := generateTestRegistry(5, 5)
	for i, service := range registry.Services {
		service.Added = service.Added.Add(-15 * time.Hour)
		registry.Services[i] = service
	}
	registry.getAll("", false)
	if length := len(registry.Services); length != 25 {
		t.Fatalf("failed to propagate registry, got length: %d", length)
	}
	registry.SetKeep(14 * time.Hour)
	registry.getAll("", false)
	if length := len(registry.Services); length != 0 {
		t.Fatalf("expected empty list, got length: %d", length)
	}
}
