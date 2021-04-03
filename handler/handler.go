// Copyright 2021 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package handler

import (
	"errors"
	"net/http"
)

// Handler is a http.handler for given matcher
type Handler struct {
	HTTPHandler http.Handler

	segments []string
	params   map[string]int
}

const (
	separator        = '/'
	paramInitialChar = ':'
)

// New returns a new Handler
func New(path string, h http.Handler) (*Handler, error) {
	segments := make([]string, 0)
	params := make(map[string]int)

	if len(path) < 1 {
		return nil, errors.New("path can't be empty")
	}
	if path[0] != '/' {
		return nil, errors.New("path must start with '/' char")
	}
	if h == nil {
		return nil, errors.New("handler can't be nil")
	}
	if len(path) == 1 {
		segments = []string{""}
	}

	for i := 1; i < len(path); i++ {
		start := i
		for i < len(path) {
			if path[i] == separator {
				break
			}
			i++
		}
		segment := path[start:i]
		if segment[0] == paramInitialChar {
			params[segment[1:]] = len(segments)
		}
		segments = append(segments, segment)
	}

	return &Handler{
		HTTPHandler: h,
		segments:    segments,
		params:      params,
	}, nil
}

// Params extracts and return params from the given path segments
func (h *Handler) Params(segments []string) map[string]string {
	params := make(map[string]string, len(h.params))
	for name, index := range h.params {
		params[name] = segments[index]
	}
	return params
}

// Segments returns segments
func (h *Handler) Segments() []string {
	return h.segments
}
