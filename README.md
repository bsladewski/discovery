# discovery

Package discovery implements a service registry for tracking the location of
distributed microservices.

## Installation

`go get github.com/bsladewski/discovery`

## Command Usage

`$ discovery`

This starts a discovery service on default port 80.

### Specifying Port

`$ discovery -port 58585`

### Adding Basic Auth

`$ discovery -user "username" -pass "password"`

### Using TLS

`$ discovery -cert "/path/to/cert" -key "/path/to/key"`

### Logging to File

`$ discovery -log "/path/to/logfile"`

## Examples

The following examples show how to use the discovery package in client code.
The server and client types also provide constructors for using TLS. A server
or client constructed using TLS is used in the same way.

### Server

```go
server := discovery.NewRandomServer(port, auth)
err := server.Run()
```

- `port` a port number.
- `auth` an instance of the `Authenticator` type.

By default, the server will consider a service active for one minute without
renewal and keep the service in the registry for twelve hours. To change this
use the `SetTimeout` and `SetKeep` functions:

```go
server.SetTimeout(time.Hour)
server.SetKeep(24*time.Hour)
```

The `Server` type also provides a `Shutdown` function for concurrent use:

```go
...
go server.Run()
...
ctx, cancel := context.WithTimeout(context.Background, 5*time.Second)
defer cancel()
err := server.Shutdown(ctx)
```

### Authenticator

An `Authenticator` is passed into the constructor for a `Server` to define how
http authentication should be handled.

To disable http authentication use `discovery.NullAuthenticator`.

To construct an `Authenticator` that performs Basic Auth use:

```go
auth := discovery.NewBasicAuthenticator(user, pass)
```

For a custom `Authenticator` the `func(token string) bool` prototype can be used:

```go
auth := func(token string) bool {
    ...
}
```

### Client

```go
client, err := discovery.NewClient(discoveryHost, discoveryAuth, timeout)
```

- `discoveryHost` the target discovery service.
- `discoveryAuth` the token for the http Authentication header.
- `timeout` the timeout for http requests.

A `Client` or `RegistryClient` constructor will make a request against the
discovery service `Ping` endpoint before returning the client. If it is unable
to make a request against the discovery service it will return an error instead.

To get the location of a particular service:

```go
host, err := client.Discover("serviceName")
```

To get a list of services of a given type:

```go
services, err := client.List("serviceName")
```

Passing the empty string to `List` will list all services in the registry.

### Registry Client

```go
registryClient, err := discovery.NewRegistryClient(myName, myHost, discoveryHost, discoveryAuth, timeout)
```

- `myName` the name of the client service.
- `myHost` the host of the client service.
- `discoveryHost` the target discovery service.
- `discoveryAuth` the token for the http Authentication header.
- `timeout` the timeout for http requests.

To register the service with the registry:

```go
err := registryClient.Register()
```

`Register`, however, is typically not called directly. The intended use of the
`RegistryClient` is to call `Auto` on service start, and `Deregister` on service
termination.

For automatic registration on an interval use:

```go
registryClient.Auto(interval)
```

- `interval` how often the service should renew its registration.

To deregister the service and stop automatic registration:

```go
err := registryClient.Deregister()
```

### Registry

```go
// Service holds information about a service as well as the last time the
// service was renewed.
type Service struct {
	Name  string    `json:"name"`
	Host  string    `json:"host"`
	Added time.Time `json:"added"`
}
```

Information about a service is stored in the `Service` type. This type is used
internally by a service registry and a list of type `Service` is returned by the
client `List` function.

```go
// Registry holds host names for services by name.
type Registry interface {
	Add(service Service)              // Add adds or updates a service to this registry.
	Remove(service Service)           // Remove removes a service from this registry.
	Get(name string) (Service, error) // Get gets the specified service.
	List(name string) []Service       // List gets all services filtered by name.
	SetTimeout(timeout time.Duration) // SetTimeout updates the timeout duration.
	SetKeep(timeout time.Duration)    // SetKeep updates the keep duration.
}
```

A `Registry` backs the discovery service. The implementation included with the
discovery package is the `RandomRegistry` that load balances by choosing a
random replicant where more than one exists. The `RandomRegistry` is used by
`NewRandomServer`.

To use a custom load balancing algorithm an alternative implementation for
`Registry` can be used to construct a `Server` through the `NewServer` function.