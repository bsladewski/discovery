# discovery

Package discovery implements a service registry for tracking the location of distributed microservices.

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
The server and client types also provide constructors for using TLS, these
types are used in the same way but take paths to key and/or certificate files
on construction.

### Server

```go
server := discovery.NewServer(80, auth)
err := server.Run()
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

To disable http authentication use `discovery.NullAuthenticator`.

To construct an `Authenticator` using Basic Auth with use:

```go
auth := discovery.NewBasicAuthenticator("username", "password")
```

For a custom `Authenticator` the `func(token string) bool` prototype can be used:

```go
auth := func(token string) bool {
    ...
}
```

### Client

```go
client, err := discovery.NewClient("http://localhost", "authToken", 10*time.Second)
```

- `"http://localhost"` the target discovery service.
- `"authToken"` the token for the http Authentication header.
- `10*time.Second` the timeout for the http request.

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
registryClient, err := discovery.NewRegistryClient("serviceName", "http://localhost:58580", "http://localhost", "authToken", 10*time.Second)
```

- `"serviceName"` the name of the client service.
- `"http://localhost:58580"` the host of the client service.
- `"http://localhost"` the target discovery service.
- `"authToken"` the token for the http Authentication header.
- `10*time.Second` the timeout for the http request.

To register the service with the registry:

```go
err := registryClient.Register()
```

For automatic registration on an interval use:

```go
registryClient.Auto(30*time.Second)
```

- `30*time.Second` how often the service should renew its registration.

To deregister the service and stop automatic registration:

```go
err := registryClient.Deregister()
```
