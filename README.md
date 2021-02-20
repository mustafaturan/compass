# 🧭 Compass

[![GoDoc](https://godoc.org/github.com/mustafaturan/compass?status.svg)](https://godoc.org/github.com/mustafaturan/compass)
[![Build Status](https://travis-ci.com/mustafaturan/compass.svg?branch=master)](https://travis-ci.com/mustafaturan/compass)
[![Coverage Status](https://coveralls.io/repos/github/mustafaturan/compass/badge.svg?branch=master)](https://coveralls.io/github/mustafaturan/compass?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/mustafaturan/compass)](https://goreportcard.com/report/github.com/mustafaturan/compass)
[![GitHub license](https://img.shields.io/github/license/mustafaturan/compass.svg)](https://github.com/mustafaturan/compass/blob/master/LICENSE)

Compass is a HTTP server router with middleware support.

## Installation

Via go packages:
```go get github.com/mustafaturan/compass```

## Version

Compass is using semantic versioning rules for versioning.

## Usage

### Configure

Init router with options
```go
import (
	"net/http"
	"github.com/mustafaturan/compass"
)

...

router := compass.New(
	// register allowed schemes, if not specified default allows all
	compass.WithSchemes("http", "https"),

	// register allowed hostnames, if not specified default allows all
	compass.WithHostnames("localhost", "yourdomain.com"),

	// register not found handler
	compass.WithHandler(404, http.NotFoundHandler()),

	// register internal server error handler
	compass.WithHandler(500, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if err := recover(); err != nil {
			http.Error(rw,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}
	})),

	// NOTE: Only allowed to register directly for 404 and 500 status code
	// handlers to enable a way to handle most common error cases easily

	// register interceptors(middlewares)
	compass.WithInterceptors(interceptor1, interceptor2, interceptor3),
)
```

### Routes

```go
// pass options to New function as in the configure section
// there are currently 3 options available
// scheme registry: compass.WithSchemes("http", "https"), default: "*"
// hostname registry: compass.WithHostnames("localhost"), default: "*"
// interceptor registry: compass.WithInterceptor(interceptor1), no default
router := compass.New()

// <Method>: Get, Head, Post, Put, Patch, Delete, Connect, Options, Trace
// <path>: routing path with params
// <http.Handler>: any implementation of net/http.Handler
router.<Method>(<path>, <http.Handler>)
```

**Path examples:**

```
"/posts" -> params: nil
"/posts/:id" -> params: id
"/posts/:id/comments" -> params: id
"/posts/:id/comments/:commentID/likes" -> params: id, commentID
"/posts/:id/reviews" -> params: id
"/posts/:id/reviews/:reviewID" -> params: id, reviewID
```

### Serving

```go
// init router with desired options
router := compass.New()

// register routes with any http handler
router.Get("/tests/:id", func(rw http.ResponseWriter, req *http.Request) {
	params := compass.Params(req.Context())
	response := fmt.Sprintf("tests endpoint called with %s", params["id"])
	rw.Write([]byte())
})

// init an HTTP server as usual and pass the router instance as handler
srv := &http.Server{
	Handler:      router, // compass.Router
	Addr:         "127.0.0.1:8000",
	// Set server options as usual
	WriteTimeout: 15 * time.Second,
	ReadTimeout:  15 * time.Second,
}

// start serving
log.Fatal(srv.ListenAndServe())
```

### Accessing Routing Params

Compass writes routing params to request's context, to access params, `Params`
function can be used:

```go
// returns map[string]string
params := compass.Params(ctx)
```

### Interceptors

Interceptors are basically middlewares. The interceptors are compatible with
popular muxer `gorilla/mux` library middlewares which implements also the same
`Middleware(handler http.Handler) http.Handler` function. So, `gorilla/mux`
middlewares can directly be used as `Interceptor`.

**Creating a new interceptor:**

Option 1) Implement the `interceptor.Interceptor` interface:
```go
import (
	"github.com/mustafaturan/compass/interceptor"
	...
)

...

type myinterceptor struct {}

func (i *myinterceptor) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Middleware", "compass")
		h.ServeHTTP(rw, req)
	})
}

// init router with interceptor/middleware
i := &myinterceptor{}
router := compass.New(router.WithInterceptors(i))
```

Option 2) Use home-ready implementation of Interceptor function
`interceptor.Func`

```go
import (
	"github.com/mustafaturan/compass/interceptor"
	...
)
...

myMiddleware := interceptor.Func(func(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Header().Set("X-Middleware", "compass")
		h.ServeHTTP(rw, req)
	})
})

// init router with interceptor/middleware
router := compass.New(router.WithInterceptors(myMiddleware))
```

## Contributing

All contributors should follow [Contributing Guidelines](CONTRIBUTING.md) before
creating pull requests.

## Credits

[Mustafa Turan](https://github.com/mustafaturan)

## Disclaimer

This is just a hobby project that is used in some fun projects. There is no
warranty even on the bug fixes but please file if you find any. Please use at
your own risk.

## License

Apache License 2.0

Copyright (c) 2021 Mustafa Turan

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
