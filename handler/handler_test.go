package handler

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		path       string
		segments   []string
		params     map[string]int
		errMessage string
	}{
		{
			path:     "/posts",
			segments: []string{"posts"},
			params:   make(map[string]int),
		},
		{
			path:     "/posts/:id",
			segments: []string{"posts", ":id"},
			params:   map[string]int{"id": 1},
		},
		{
			path:     "/posts/:id/comments",
			segments: []string{"posts", ":id", "comments"},
			params:   map[string]int{"id": 1},
		},
		{
			path:     "/posts/:id/comments/:commentID",
			segments: []string{"posts", ":id", "comments", ":commentID"},
			params:   map[string]int{"id": 1, "commentID": 3},
		},
		{
			path:       "",
			errMessage: "path can't be empty",
		},
		{
			path:       "x",
			errMessage: "path must start with '/' char",
		},
	}

	for _, test := range tests {
		h, err := New(test.path, testHTTPHandler{})
		t.Run("has correct error message", func(t *testing.T) {
			if test.errMessage != "" && err.Error() != test.errMessage {
				t.Fatalf(
					"must result with err(%s) for path %s but got err(%s)",
					test.errMessage,
					test.path,
					err.Error(),
				)
			}
		})
		if err != nil {
			continue
		}

		t.Run("has correct segmentation", func(t *testing.T) {
			if !reflect.DeepEqual(test.segments, h.Segments()) {
				t.Fatalf("want: %+v, got: %+v", test.segments, h.Segments())
			}
		})
		t.Run("has correct params", func(t *testing.T) {
			if !reflect.DeepEqual(test.params, h.params) {
				t.Fatalf("want: %+v, got: %+v", test.params, h.params)
			}
		})
	}

	t.Run("without handler", func(t *testing.T) {
		_, err := New("/valid", nil)
		if err.Error() != "handler can't be nil" {
			t.Fatalf("should not allow initialization with nil http handler")
		}
	})
}

func TestParams(t *testing.T) {
	tests := []struct {
		path            string
		segments        []string
		requestSegments []string
		params          map[string]string
	}{
		{
			path:            "/",
			segments:        []string{},
			requestSegments: []string{},
			params:          make(map[string]string),
		},
		{
			path:            "/:name",
			segments:        []string{":name"},
			requestSegments: []string{"test"},
			params:          map[string]string{"name": "test"},
		},
		{
			path:            "/posts",
			segments:        []string{"posts"},
			requestSegments: []string{"posts"},
			params:          make(map[string]string),
		},
		{
			path:            "/posts/:id",
			segments:        []string{"posts", ":id"},
			requestSegments: []string{"posts", "1"},
			params:          map[string]string{"id": "1"},
		},
		{
			path:            "/posts/:id/comments",
			segments:        []string{"posts", ":id", "comments"},
			requestSegments: []string{"posts", "1", "comments"},
			params:          map[string]string{"id": "1"},
		},
		{
			path:            "/posts/:id/comments/:commentID",
			segments:        []string{"posts", ":id", "comments", ":commentID"},
			requestSegments: []string{"posts", "1", "comments", "99"},
			params:          map[string]string{"id": "1", "commentID": "99"},
		},
	}

	for _, test := range tests {
		t.Run("build correct params", func(t *testing.T) {
			h, _ := New(test.path, testHTTPHandler{})
			params := h.Params(test.requestSegments)
			if !reflect.DeepEqual(test.params, params) {
				t.Fatalf("want: %+v, got: %+v", test.params, params)
			}
		})
	}
}

type testHTTPHandler struct{}

func (h testHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
}
