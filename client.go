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
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Client an http client to the discovery server.
type Client struct {
	http.Client
	host  string
	token string
}

// Discover gets the host of the target service by name or an error.
func (client *Client) Discover(name string) (string, error) {
	values := url.Values{}
	values.Add("name", name)
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "discover"))
	uri.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("Authorization", client.token)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf(string(body))
	}
	service := Service{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&service)
	if err != nil {
		return "", err
	}
	return service.Host, nil
}

// List lists all services filtered by name.
func (client *Client) List(name string) ([]Service, error) {
	values := url.Values{}
	values.Add("name", name)
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "list"))
	uri.RawQuery = values.Encode()
	req, err := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("Authorization", client.token)
	resp, err := client.Do(req)
	if err != nil {
		return []Service{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []Service{}, err
		}
		return []Service{}, fmt.Errorf(string(body))
	}
	services := struct {
		Services []Service `json:"services"`
	}{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&services)
	if err != nil {
		return []Service{}, err
	}
	return services.Services, nil
}

// Ping pings the discovery service.
func (client *Client) Ping() error {
	uri, _ := url.Parse(fmt.Sprintf("%s/%s", client.host, "ping"))
	req, err := http.NewRequest("GET", uri.String(), nil)
	req.Header.Set("Authorization", client.token)
	resp, err := client.Do(req)
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

// NewClient returns a discovery server client.
func NewClient(host, token string, timeout time.Duration) (*Client, error) {
	client := &Client{
		http.Client{
			Timeout: timeout,
		},
		host,
		token,
	}
	err := client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err.Error())
	}
	return client, nil
}

// NewTLSClient returns an encrypted discovery server client.
func NewTLSClient(host, token, certFile string,
	skipVerify bool, timeout time.Duration) (*Client, error) {
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
	client := &Client{
		http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipVerify,
					RootCAs:            certs,
				},
			},
		},
		host,
		token,
	}
	err = client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err.Error())
	}
	return client, nil
}
