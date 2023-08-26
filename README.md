# API Gateway

Lightweight API Gateway written in Go.

## Config

The main way of configuring the Gateway is done by the `config.json` file. You can always find the latest state of the config in `config.json.example` file, even if this document is not updated.

A basic config looks something like this:
```json
{
  "address": 3100,							// This is the HTTP address where the Gateway will listen at.
  "productionLevel": 0,					// The state of the software. 0 stands for development and 1 for production.
  "middlewaresEnabled": 1,			// Flag for enabling global and custom middlewares. In case of 0 they are disabled, case of 1 they are enabled.
  "healthCheckInterval": "1m",	// The interval of the health check frequency, later discussed. 
  "secretKey": "",							// SecretKey later discussed.
	"loggerConfig": {
    "disabledLoggers": [				// If you dont want to see loggers, you can disable it here.
			// "info",
			// "warning",
      "error"
		]
  },
  "services": [									// All the registered services, later discussed.
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
  "protocol": "http", 			// Can be "http" or "https".
  "host": "localhost",			// Any hostname without without protocol.
  "name": "exampleService", // The name of the service, must be unique.
  "port": "3001",						// The port where the service is avaiable.
  "prefix": "/api/test",		// The routing prefix.
  "timeOutSec": 5						// How long should a request go on before timeouting.
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
