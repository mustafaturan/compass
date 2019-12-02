// Copyright 2019 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package middleware

import (
	"net/http"
)

// Middleware interface is anything which implements a MiddlewareFunc
type Middleware interface {
	Next(handler http.Handler) http.Handler
}

// Func is a function which receives an http.Handler and returns another
// http.Handler. Typically, the returned handler is a closure which does
// something with the http.ResponseWriter and http.Request passed to it, and
// then calls the handler passed as parameter to the Func.
type Func func(http.Handler) http.Handler

// Next allows Func to implement middleware interface
func (fn Func) Next(handler http.Handler) http.Handler {
	return fn(handler)
}
