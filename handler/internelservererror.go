// Copyright 2019 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package handler

import (
	"net/http"
)

// InternalServerError implements http.Handler
type InternalServerError struct{}

// ServeHTTP implements http handler func for http.Handler interface
func (h InternalServerError) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := recover(); err != nil {
		http.Error(rw,
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
}
