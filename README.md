# API Gateway

Lightweight API Gateway written in Go.

## Config

The main way of configuring the Gateway is done by the `config.json` file. You can always find the latest state of the config in `config.json.example` file, even if this document is not updated.

A basic config looks something like this:
```json
{
  "address": 3100,
  "productionLevel": 0,
  "middlewaresEnabled": 1,
  "healthCheckInterval": "1m",
  "secretKey": "",
	"loggerConfig": {
    "disabledLoggers": [
      "error"
		]
  },
  "services": [
    {
      "protocol": "http",
      "name": "testService",
      "host": "localhost",
      "port": "3001",
      "prefix": "/api/test",
      "timeOutSec": 5
    }
  ]
}
```

## Features

### Creating a new instance

There are two main ways to create a new instance. The first one is the more general, that reads in the config file, then initiates the instance by given details. The only parameter it requires is the relative path for the config file. It returns a pointer to the new instance and an error, if there is any.

```go
gw, err := gateway.NewFromConfig("./example.config.json")
if err != nil {
	fmt.Println("gateway create err: %v\n", err)
	os.Exit(1)
}
```

The other way is the more traditional one, where we can initiate a new instance programatically with the decoractor pattern. It does not return any error, only the pointer to the new Gateway instance.

```go
func main() {
	gw := gateway.New(
		gateway.WithAddress(8000),
		gateway.WithHealthCheckFrequency(5*time.Second),
	)

	gw.Start()
}
```

### Custom endpoints

It is possible for the Gateway to acts a router itself, by registering routes with handlers. 

The router is REST compatible, which means, is supports selection based upon HTTP methods, and wildcard path params. Eg.: GET /api/foo/{id} and DELETE /api/foo/{id}

Example:

```go
gw.Get("/api/foo/{id}", func(ctx *gateway.Context) {
	id := ctx.GetParam(id)

	type response struct {
		param string
	}

	res := &response{
		param: id,
	}

	ctx.SendJson(res)
})

```

### Endpoint middlewares

As it is mentioned earlier, it is possible to register custom HTTP endpoints. Also there is a way to attach middlewares to each one. Every given middleware function is attached to the endpoint as a `pre runner`, which means, the middleware functions run before the execution of the handler itself. The sequence of the middleware chain is the same as the registrations order.

Example:
```go
mw1 := func(ctx *gateway.Context, next gateway.HandlerFunc) {
	// Some work or auth to do.
	// ...

	// Then we call the next in the sequence.
	next(ctx)
}

mw2 := func(ctx *gateway.Context, next gateway.HandlerFunc) {
	// Some other work or auth to do.
	// ...

	// Then we call the next in the sequence.
	next(ctx)
}

gw.Get("/api/foo/bar", func(ctx *gateway.Context) {
	ctx.SendOk()
}).RegisterMiddlewares(mw1, mw2)
```

### Global middlewares

Besides the middlewares that are attached to specific endpoints, we can register global middlewares, which takes a normal middlewarefunc – as mentioned before – and a matcherfunc as a necessity. The matcher takes the Context as a parameter and returns a boolean, that indicates for a match. Eg.: it could be a match for URL.

Example – the MW is only called if the url contains `foo`:

```go
var matcher = func(ctx *gateway.Context) bool {
	return strings.Contains(ctx.GetFullUrl(), "foo")
}

gw.RegisterMiddleware(func(ctx *gateway.Context, next gateway.HandlerFunc) {
	// Some other work or auth to do.
	// ...

	// Then we call the next in the sequence.
	next(ctx)
}, matcher)
```

### Logging to file 

As well as normal logging to stdout and stderr, it is enabled by deafult to write the same logs to persistent files, which date stamps.

### Service registry

The main feature of any API Gataway is the abilitiy to handle traffic to and between different services. In this gateway the routing is based upon prefixing.

Every service's config must be follow this rule in the `config.json` file:

```json
{
  "protocol": "http",
  "host": "localhost",
  "name": "exampleService",
  "port": "3001",
  "prefix": "/api/test",
  "timeOutSec": 5
}
```

After the Gateway is up and running, it will make requests to the registered services – as a heartbeat – periodically. Each registered services must have a public REST endpoint: `GET /api/status/health-check`. It should only respond with HTTP 200. Any other status code or timeout will be acknowledged as the given service is down.

There is another way to signal the Gateway that one service is up, is by making a POST request as the following. The url be: `/api/system/services/update`.

The request body :
```json
{
	"serviceName": "exampleService"
}
```

To ensure that this the request is done by an authorized service, the following header must be present, or the request is not proccessed:

```plain
'X-GATEWAY-KEY': $HASHVALUE
```

where the `$HASHVALUE` is calculated by the following method. Lets take body of the request, stringify it, and compact it – remove all unnecessary whitespaces and newline characters. Then append the common secret key to the and. eq: `{"serviceName":"exampleService"}exampleSecretKey`. Then, you must make a hash with SHA256 algorithm and you are done.

If a service is down you are trying to access it, the Gateway would return an HTTP 503 error, as expected.

There is way to get some information about the inner state of the Gateway and service. You have to make a POST request to: `/api/system/services/info`. The body must be an empty object: `{}`, and the it should include the appended secret key and also the header aswell.


### gRPC proxy

With version `v0.4.0` a gRPC proxy is introduced int the Gateway. In order to start it, simply have to add this into the main configuration file of the Gateway:

```json
"grpcProxy": {
  "address": 3000
}
```

This addition to the config, will start a gRPC sever listening at the given port. This will proxy all the gRPC calls between the services in the cluster. 

For now, this gRPC proxy only supports interservice communication, so from the outside only REST calls are supported.

To mark a service as gRPC compatible service, only have to modify the config of the given service as below:

```json
"services": [
  {
    "serviceType": 1,
    "protocol": "http",
    "name": "exampleService",
    "host": "localhost",
    "port": "3001",
    "prefix": "/example.ExampleService"
  }
]
```

Where the `serviceType` must take the value `1`, and the prefix should be a unique part of the gRPC service `FullMethodName`. 

To identify this, you have to look inside the generated `*._grpc.pb.go` file. There you would find something like this:

```go
const (
	ExampleService_GetMessage_FullMethodName = "/example.ExampleService/GetMessage"
)
```

Probably, there is more than one `FullMethodName` in your file, but the package description will be the same.

```go
const (
	ExampleService_GetMessage_FullMethodName 	= "/example.ExampleService/GetMessage"
	OtherService_GetMessage_FullMethodName 		= "/example.OtherService/GetMessage"
)
```

In the case of the latter example, the prefix should be `/example`. Every gRCP proxy call will make a lookup inside the `Service registry`, and find the best fit, due to the longest match in the given prefix.
