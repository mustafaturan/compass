// Copyright 2021 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package compass

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	chandler "github.com/mustafaturan/compass/handler"
	cinterceptor "github.com/mustafaturan/compass/interceptor"
	cmatcher "github.com/mustafaturan/compass/matcher"
)

// Router is an internal router for HTTP routing with interceptor support
type Router interface {
	http.Handler

	// Get registers handler for GET method
	Get(path string, handler http.Handler) error
	// Head registers handler for HEAD method
	Head(path string, handler http.Handler) error
	// Post registers handler for POST method
	Post(path string, handler http.Handler) error
	// Put registers handler for PUT method
	Put(path string, handler http.Handler) error
	// Patch registers handler for PATCH method
	Patch(path string, handler http.Handler) error
	// Delete registers handler for DELETE method
	Delete(path string, handler http.Handler) error
	// Connect registers handler for CONNECT method
	Connect(path string, handler http.Handler) error
	// Options registers handler for OPTIONS method
	Options(path string, handler http.Handler) error
	// Trace registers handler for TRACE method
	Trace(path string, handler http.Handler) error
}

// router is an implementation of Router
type router struct {
	interceptors []cinterceptor.Interceptor
	matcher      *cmatcher.Matcher

	// Schemes allows access to the provided schemes only
	// The default value catches `http` and `https` schemes
	schemes map[string]struct{}

	// Hostnames allows access to the provided hostnames only
	// The default value catches all hostnames (`*`)
	hostnames map[string]struct{}

	// NotFound http handler
	notfound http.Handler

	// InternalServerError http handler for panic recovery
	internalservererror http.Handler
}

// Option is a router option
type Option func(*router) error

type ctxKey int8

const (
	// CtxParams params context key
	CtxParams = ctxKey(0)

	// matchall char to match any hostname or scheme
	matchall = "*"
)

// New returns a new Router with default handlers
func New(options ...Option) (Router, error) {
	r := &router{
		interceptors:        make([]cinterceptor.Interceptor, 0),
		matcher:             cmatcher.New(),
		schemes:             map[string]struct{}{matchall: {}},
		hostnames:           map[string]struct{}{matchall: {}},
		notfound:            http.NotFoundHandler(),
		internalservererror: chandler.InternalServerError{},
	}

	for _, o := range options {
		if err := o(r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

// WithSchemes option sets allowed schemes
func WithSchemes(schemes ...string) Option {
	return func(r *router) error {
		for s := range r.schemes {
			delete(r.schemes, s)
		}

		for _, s := range schemes {
			r.schemes[s] = struct{}{}
		}
		return nil
	}
}

// WithHostnames option sets allowed hostnames
func WithHostnames(hostnames ...string) Option {
	return func(r *router) error {
		for h := range r.hostnames {
			delete(r.hostnames, h)
		}

		for _, h := range hostnames {
			r.hostnames[h] = struct{}{}
		}
		return nil
	}
}

// WithHandler option registers default handlers for NotFound and
// InternalServerError error status codes
func WithHandler(statusCode int, h http.Handler) Option {
	return func(r *router) error {
		if h == nil {
			return errors.New("handler can't be nil")
		}
		switch statusCode {
		case 404:
			r.notfound = h
		case 500:
			r.internalservererror = h
		default:
			return fmt.Errorf("can't set a default handler for status code %d", statusCode)
		}
		return nil
	}
}

// WithInterceptors appends a interceptor.Interceptor to the chain. Interceptor
// can be used to intercept or otherwise modify requests and/or responses, and
// are executed in the order that they are applied to the Router.
func WithInterceptors(interceptors ...cinterceptor.Interceptor) Option {
	return func(r *router) error {
		r.interceptors = append(r.interceptors, interceptors...)
		return nil
	}
}

// Params provides access to compass params
func Params(ctx context.Context) map[string]string {
	return ctx.Value(CtxParams).(map[string]string)
}

// ServeHTTP implements http.Handler interface with interceptors
func (r *router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer r.internalservererror.ServeHTTP(rw, req)

	var h http.Handler
	var params map[string]string

	if !r.isAllowedScheme(req.URL.Scheme) || !r.isAllowedHostname(req.URL.Hostname()) {
		h, params = r.notfound, make(map[string]string)
	} else {
		h, params = r.match(req)
	}

	// Attach params to request with context
	ctx := context.WithValue(req.Context(), CtxParams, params)
	req = req.WithContext(ctx)

	for i := len(r.interceptors) - 1; i >= 0; i-- {
		h = r.interceptors[i].Middleware(h)
	}

	h.ServeHTTP(rw, req)
}

// Get registers handler for GET method
func (r *router) Get(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodGet, path, handler)
}

// Head registers handler for HEAD method
func (r *router) Head(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodHead, path, handler)
}

// Post registers handler for POST method
func (r *router) Post(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodPost, path, handler)
}

// Put registers handler for PUT method
func (r *router) Put(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodPut, path, handler)
}

// Patch registers handler for PATCH method
func (r *router) Patch(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodPatch, path, handler)
}

// Delete registers handler for DELETE method
func (r *router) Delete(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodDelete, path, handler)
}

// Connect registers handler for CONNECT method
func (r *router) Connect(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodConnect, path, handler)
}

// Options registers handler for OPTIONS method
func (r *router) Options(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodOptions, path, handler)
}

// Trace registers handler for TRACE method
func (r *router) Trace(path string, handler http.Handler) error {
	return r.registerHandler(http.MethodTrace, path, handler)
}

func (r *router) registerHandler(method, path string, handler http.Handler) error {
	h, err := chandler.New(path, handler)
	if err != nil {
		return err
	}
	return r.matcher.Register(method, h)
}

func (r *router) isAllowedHostname(hostname string) bool {
	if _, hasHostname := r.hostnames[hostname]; hasHostname {
		return true
	}
	_, hasMatchAll := r.hostnames[matchall]
	return hasMatchAll
}

func (r *router) isAllowedScheme(scheme string) bool {
	if _, hasScheme := r.schemes[scheme]; hasScheme {
		return true
	}
	_, hasMatchAll := r.schemes[matchall]
	return hasMatchAll
}

func (r *router) match(req *http.Request) (http.Handler, map[string]string) {
	path := req.URL.EscapedPath()
	segments := strings.Split(path[1:], "/")
	if h, found := r.matcher.Find(req.Method, segments); found {
		return h.HTTPHandler, h.Params(segments)
	}
	return r.notfound, make(map[string]string)
}
