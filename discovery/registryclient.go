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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
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
	running  bool
	shutdown chan bool
}

// Register registers the service with the discovery service.
func (client *RegistryClient) Register() error {
	raw, err := json.Marshal(client.service)
	if err != nil {
		return err
	}
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "register"))
	req, err := http.NewRequest("POST", uri.String(), bytes.NewBuffer(raw))
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

// doAuto a concurrent function to perform the automatic registration.
func (client *RegistryClient) doAuto(interval time.Duration) {
	client.running = true
	for {
		select {
		case <-client.shutdown:
			client.running = false
			return
		default:
			client.Register()
			time.Sleep(interval)
		}
	}
}

// Auto automatically registers the service with the discovery service on the
// specified interval.
func (client *RegistryClient) Auto(interval time.Duration) {
	if !client.running {
		go client.doAuto(interval)
	}
}

// Deregister deregisters the service with the discovery service. Terminates
// auto register if enabled.
func (client *RegistryClient) Deregister() error {
	if client.running {
		select {
		case client.shutdown <- true:
		default:
		}
	}
	raw, err := json.Marshal(client.service)
	if err != nil {
		return err
	}
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "deregister"))
	req, err := http.NewRequest("DELETE", uri.String(), bytes.NewBuffer(raw))
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

// NewRegistryClient returns a discovery server registry client.
func NewRegistryClient(name, host, targetHost, targetToken string,
	timeout time.Duration) (*RegistryClient, error) {
	client := &RegistryClient{
		host:     targetHost,
		token:    targetToken,
		service:  Service{Name: name, Host: host},
		shutdown: make(chan bool, 1),
	}
	client.netClient = &http.Client{
		Timeout: timeout,
	}
	err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err.Error())
	}
	return client, nil
}

// NewTLSRegistryClient returns an encryped discovery server registry client.
func NewTLSRegistryClient(name, host, targetHost, targetToken, certFile string,
	skipVerify bool, timeout time.Duration) (*RegistryClient, error) {
	client := &RegistryClient{
		host:     targetHost,
		token:    targetToken,
		service:  Service{Name: name, Host: host},
		shutdown: make(chan bool, 1),
	}
	certs, err := x509.SystemCertPool()
	if err != nil {
		certs = x509.NewCertPool()
	}
	if certFile != "" {
		pemData, err := ioutil.ReadFile(certFile)
		if err != nil {
			return nil, err
		}
		if !certs.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("failed to load specified certificate")
		}
	}
	client.netClient = &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipVerify,
				RootCAs:            certs,
			},
		},
	}
	err = client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err.Error())
	}
	return client, nil
}
