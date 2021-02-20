// Copyright 2021 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

/*
Package compass is a HTTP server router library with middleware support.

Via go packages:
go get github.com/mustafaturan/compass

Version

Compass is using semantic versioning rules for versioning.

Usage

### Configure

Init router with options

	router := compass.New(
		// register allowed schemes
		compass.WithSchemes("http", "https"),

		// register allowed hostnames
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

		// register interceptors(middlewares)
		compass.WithInterceptors(interceptor1, interceptor2, interceptor3),
	)


### Routes

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

**Path examples:**

	"/posts" -> params: nil
	"/posts/:id" -> params: id
	"/posts/:id/comments" -> params: id
	"/posts/:id/comments/:commentID/likes" -> params: id, commentID
	"/posts/:id/reviews" -> params: id
	"/posts/:id/reviews/:reviewID" -> params: id, reviewID

### Serving

	router := compass.New()

	srv := &http.Server{
		Handler:      router, // compass.Router
		Addr:         "127.0.0.1:8000",
		// Set server options as usual
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

### Accessing Routing Params

Compass writes routing params to request's context, to access params, `Params`
function can be used:

	// returns map[string]string
	params := compass.Params(ctx)

### Interceptors

Interceptors are basically middlewares. The interceptors are compatible with
popular muxer `gorilla/mux` library middlewares which implements also the same
`Middleware(handler http.Handler) http.Handler` function. So, `gorilla/mux`
middlewares can directly be used as `Interceptor`.

**Creating a new interceptor:**

Option 1) Implement the `interceptor.Interceptor` interface:

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

Option 2) Use home-ready implementation of Interceptor function
`interceptor.Func`

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

*/
package compass
