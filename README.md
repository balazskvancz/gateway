# API Gateway

Lightweight API Gateway written in Go.

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