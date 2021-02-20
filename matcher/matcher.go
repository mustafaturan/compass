// Copyright 2021 Mustafa Turan. All rights reserved.
// Use of this source code is governed by a Apache License 2.0 license that can
// be found in the LICENSE file.

package matcher

import (
	"errors"
	"net/http"

	chandler "github.com/mustafaturan/compass/handler"
)

const (
	pathvar = ":"
)

// Matcher is a modified version of Radix tree for HTTP Routing
type Matcher struct {
	nodes map[string]*node
}

type node struct {
	handler *chandler.Handler

	nodes map[string]*node
}

// New inits a new matcher
func New() *Matcher {
	return &Matcher{nodes: map[string]*node{
		http.MethodGet:     {nodes: make(map[string]*node)},
		http.MethodHead:    {nodes: make(map[string]*node)},
		http.MethodPost:    {nodes: make(map[string]*node)},
		http.MethodPut:     {nodes: make(map[string]*node)},
		http.MethodPatch:   {nodes: make(map[string]*node)},
		http.MethodDelete:  {nodes: make(map[string]*node)},
		http.MethodConnect: {nodes: make(map[string]*node)},
		http.MethodOptions: {nodes: make(map[string]*node)},
		http.MethodTrace:   {nodes: make(map[string]*node)},
	}}
}

// Find finds the top priority HTTP handler
func (m *Matcher) Find(method string, segments []string) (*chandler.Handler, bool) {
	if len(segments) == 0 {
		return m.nodes[method].handler, m.nodes[method].handler != nil
	}
	var pn node
	m.nodes[method].search(segments, 0, &pn)

	return pn.handler, pn.handler != nil
}

// Register adds a new handler for the given path
func (m *Matcher) Register(method string, h *chandler.Handler) error {
	n := m.nodes[method].insert(h.Segments(), 0)
	if n.handler != nil {
		return errors.New("path is already registered for another handler")
	}
	n.handler = h
	return nil
}

func (n *node) search(segments []string, index int, pn *node) {
	if pn.handler != nil {
		return
	}
	if n == nil {
		return
	}
	if len(segments) == index {
		pn.handler = n.handler
		return
	}

	segment := segments[index]
	n.nodes[segment].search(segments, index+1, pn)
	n.nodes[pathvar].search(segments, index+1, pn)
}

func (n *node) insert(segments []string, index int) *node {
	if len(segments) == index {
		return n
	}

	segment := segments[index]
	if len(segment) > 0 && segment[0] == pathvar[0] {
		segment = pathvar
	}
	if next, ok := n.nodes[segment]; ok {
		return next.insert(segments, index+1)
	}
	if segment != pathvar &&
		n.nodes[pathvar] != nil &&
		n.nodes[pathvar].handler != nil &&
		len(segments) == index+1 {
		return n.nodes[pathvar]
	}

	next := &node{nodes: make(map[string]*node)}
	n.nodes[segment] = next
	return next.insert(segments, index+1)
}
