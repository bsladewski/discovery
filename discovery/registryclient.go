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
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// RegistryClient an http client to the discovery service registry features.
type RegistryClient struct {
	host  string
	token string

	netClient *http.Client

	service  Service
	shutdown chan bool
}

// Register registers the service with the discovery service.
func (client *RegistryClient) Register() error {
	return nil
}

// Deregister deregisters the service with the discovery service. Terminates
// auto register if enabled.
func (client *RegistryClient) Deregister() error {
	return nil
}

// Auto automatically registers the service with the discovery service on the
// specified interval.
func (client *RegistryClient) Auto(interval time.Duration) {

}

// Ping pings the discovery service.
func (client *RegistryClient) Ping() error {
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "ping"))
	req, err := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("Authorization", client.token)
	resp, err := client.netClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(body))
	}
	return nil
}

func NewRegistryClient(name, host, targetHost, targetToken string,
	timeout time.Duration) (*RegistryClient, error) {
	return nil, nil
}

func NewTLSRegistryClient(name, host, targetHost, targetToken, certFile string,
	skipVerify bool, timeout time.Duration) (*RegistryClient, error) {
	return nil, nil
}
