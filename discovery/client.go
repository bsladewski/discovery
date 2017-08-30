// Package discovery implements a service registry for tracking the location of
// distributed microservices.
package discovery

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

// Client an http client to the discovery server.
type Client struct {
	host  string
	token string

	netClient *http.Client
}

// Discover gets the host of the target service by name or an error.
func (client *Client) Discover(name string) (string, error) {
	values := url.Values{}
	values.Add("name", name)
	uri, _ := url.Parse(path.Join(client.host, "discover"))
	uri.RawQuery = values.Encode()
	resp, err := client.netClient.Get(uri.String())
	if err != nil {
		return "", err
	}
	service := Service{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&service)
	if err != nil {
		return service.Host, nil
	}
	return service.Host, nil
}

// List lists all services filtered by name.
func (client *Client) List(name string) ([]Service, error) {
	values := url.Values{}
	values.Add("name", name)
	uri, _ := url.Parse(client.host)
	uri.RawQuery = values.Encode()
	resp, err := client.netClient.Get(uri.String())
	if err != nil {
		return []Service{}, err
	}
	services := struct {
		Services []Service `json:"services"`
	}{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&services)
	if err != nil {
		return []Service{}, nil
	}
	return services.Services, nil
}

// NewClient returns a discovery server client.
func NewClient(host, token string, timeout time.Duration) (*Client, error) {
	client := &Client{
		host:  host,
		token: token,
	}
	client.netClient = &http.Client{
		Timeout: timeout,
	}
	_, err := client.List("")
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NewTLSClient returns an encrypted discovery server client.
func NewTLSClient(host, token, certFile string,
	skipVerify bool, timeout time.Duration) (*Client, error) {
	client := &Client{
		host:  host,
		token: token,
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
		certs.AppendCertsFromPEM(pemData)
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
	_, err = client.List("")
	if err != nil {
		return nil, err
	}
	return client, nil
}
