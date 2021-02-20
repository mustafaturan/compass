// Copyright 2021 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package interceptor

import (
	"net/http"
)

// Interceptor interface is anything which implements a interceptor.Func
type Interceptor interface {
	Middleware(handler http.Handler) http.Handler
}

// Func is a function which receives an http.Handler and returns another
// http.Handler. Typically, the returned handler is a closure which does
// something with the http.ResponseWriter and http.Request passed to it, and
// then calls the handler passed as parameter to the Func.
type Func func(http.Handler) http.Handler

// Middleware allows Func to implement interceptor interface
func (fn Func) Middleware(handler http.Handler) http.Handler {
	return fn(handler)
}
