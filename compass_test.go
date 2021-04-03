package compass

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	chandler "github.com/mustafaturan/compass/handler"
)

func TestNew(t *testing.T) {
	notfoundHandler := fakeHandler{}
	tests := []struct {
		options                     []Option
		expectedSchemes             map[string]struct{}
		expectedHostnames           map[string]struct{}
		expectedNotfound            http.Handler
		expectedInternalservererror http.Handler
		err                         error
	}{
		{
			options: []Option{
				WithSchemes("http", "https"),
				WithHostnames("*"),
				WithHandler(404, notfoundHandler),
			},
			expectedSchemes:             map[string]struct{}{"http": {}, "https": {}},
			expectedHostnames:           map[string]struct{}{"*": {}},
			expectedNotfound:            notfoundHandler,
			expectedInternalservererror: chandler.InternalServerError{},
		},
		{
			options: []Option{
				WithSchemes("http", "https"),
				WithHostnames("*"),
				WithHandler(401, fakeHandler{}),
			},
			expectedSchemes:   map[string]struct{}{"http": {}, "https": {}},
			expectedHostnames: map[string]struct{}{"*": {}},
			err:               errors.New("can't set a default handler for status code 401"),
		},
	}

	for _, test := range tests {
		ri, err := New(test.options...)
		if test.err != nil {
			t.Run("err match", func(t *testing.T) {
				if test.err.Error() != err.Error() {
					t.Fatalf("errors does not match: '%s'", err.Error())
				}
			})
			continue
		}

		r := ri.(*router)

		t.Run("has correct schemes registered", func(t *testing.T) {
			if !reflect.DeepEqual(test.expectedSchemes, r.schemes) {
				t.Fatalf("config schemes don't match")
			}
		})

		t.Run("has correct hostnames registered", func(t *testing.T) {
			if !reflect.DeepEqual(test.expectedHostnames, r.hostnames) {
				t.Fatalf("config hostnames don't match")
			}
		})

		t.Run("has correct not found http handler", func(t *testing.T) {
			if !reflect.DeepEqual(test.expectedNotfound, r.notfound) {
				t.Fatalf("not found handler does not match")
			}
		})

		t.Run("has correct internal server error http handler", func(t *testing.T) {
			if !reflect.DeepEqual(test.expectedInternalservererror, r.internalservererror) {
				t.Fatalf("internal server error handler does not match")
			}
		})
	}
}

func TestWithSchemes(t *testing.T) {
	tests := []struct {
		schemes  []string
		expected map[string]struct{}
	}{
		{[]string{"file", "ftp"}, map[string]struct{}{"file": {}, "ftp": {}}},
	}

	r := &router{schemes: map[string]struct{}{}}

	for _, test := range tests {
		err := WithSchemes(test.schemes...)(r)

		if err != nil {
			t.Fatalf("registration of schemes with %+v should not return error", test.schemes)
		}

		if !reflect.DeepEqual(test.expected, r.schemes) {
			t.Fatalf("registration of schemes hasn't set the schemes correctly %+v", test.schemes)
		}
	}
}

func TestHostnames(t *testing.T) {
	tests := []struct {
		hostnames []string
		expected  map[string]struct{}
	}{
		{[]string{"example.com", "test.com"}, map[string]struct{}{"example.com": {}, "test.com": {}}},
	}

	r := &router{hostnames: map[string]struct{}{}}

	for _, test := range tests {
		err := WithHostnames(test.hostnames...)(r)

		if err != nil {
			t.Fatalf("registration with %+v should not return error", test.hostnames)
		}

		if !reflect.DeepEqual(test.expected, r.hostnames) {
			t.Fatalf("registration of hostnames hasn't set the hostnames correctly %+v", test.hostnames)
		}
	}
}

func TestWithHandler(t *testing.T) {
	tests := []struct {
		statusCode  int
		handlerFunc http.Handler
		err         string
	}{
		{404, nil, "handler can't be nil"},
		{404, http.NotFoundHandler(), ""},
		{500, chandler.InternalServerError{}, ""},
	}

	r := &router{}

	for _, test := range tests {
		err := WithHandler(test.statusCode, test.handlerFunc)(r)

		if err != nil && err.Error() != test.err {
			t.Fatalf("registration with %+v must return err with correct message", test.handlerFunc)
		}
		if err == nil && test.err != "" {
			t.Fatalf("registration with %+v must return err", test.handlerFunc)
		}
	}
}

func TestWithInterceptors(t *testing.T) {
	r := &router{}
	first, second := &fakeInterceptor{"first"}, &fakeInterceptor{"second"}
	_ = WithInterceptors(first, second)(r)
	t.Run("adds interceptors in correct order", func(t *testing.T) {
		if r.interceptors[0] != first || r.interceptors[1] != second {
			t.Fatalf("interceptors registered in incorrect order")
		}
	})
	t.Run("has correct number of interceptors", func(t *testing.T) {
		if len(r.interceptors) != 2 {
			t.Fatalf("must register exactly 2 interceptors")
		}
	})
}

func TestParams(t *testing.T) {
	expected := map[string]string{"test": "val"}
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxParams, expected)
	val := Params(ctx)
	if !reflect.DeepEqual(expected, val) {
		t.Fatalf("vals couldn't be extracted")
	}
}

func TestServeHTTP(t *testing.T) {
	tests := []struct {
		handler    http.Handler
		path       string
		reqURL     string
		schemes    []string
		hostnames  []string
		statusCode int
		params     map[string]string
	}{
		{
			handler:    fakeHandler{"ok"},
			path:       "/",
			reqURL:     "https://example.com/",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"example.com"},
			statusCode: 200,
			params:     make(map[string]string),
		},
		{ // hostname, scheme, path registered
			handler:    fakeHandler{"ok"},
			path:       "/posts",
			reqURL:     "https://example.com/posts",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"example.com"},
			statusCode: 200,
			params:     make(map[string]string),
		},
		{ // hostname, scheme, path registered
			handler:    fakeHandler{"ok"},
			path:       "/posts",
			reqURL:     "https://example.com/posts",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"www.example.com", "example.com"},
			statusCode: 200,
			params:     make(map[string]string),
		},
		{ // hostname, scheme, path registered with interceptor
			handler:    fakeHandler{"ok"},
			path:       "/posts",
			reqURL:     "https://example.com/posts",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"www.example.com", "example.com"},
			statusCode: 200,
			params:     make(map[string]string),
		},
		{ // hostname, scheme, path registered with interceptor with params
			handler:    fakeHandler{"ok"},
			path:       "/posts/:id",
			reqURL:     "https://example.com/posts/1",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"www.example.com", "example.com"},
			statusCode: 200,
			params:     map[string]string{"id": "1"},
		},
		{ // match-all hostname, match-all scheme
			handler:    fakeHandler{"ok"},
			path:       "/posts/:id",
			reqURL:     "https://example.com/posts/1",
			schemes:    []string{"*"},
			hostnames:  []string{"*"},
			statusCode: 200,
			params:     map[string]string{"id": "1"},
		},
		{ // hostname not registered
			handler:    fakeHandler{"na"},
			path:       "/posts",
			reqURL:     "https://example.com/posts",
			schemes:    []string{"https", "http"},
			hostnames:  []string{"www.example.com"},
			statusCode: 404,
			params:     make(map[string]string),
		},
		{ // scheme not registered
			handler:    fakeHandler{"na"},
			path:       "/posts",
			reqURL:     "https://example.com/posts",
			schemes:    []string{"http"},
			hostnames:  []string{"example.com"},
			statusCode: 404,
			params:     make(map[string]string),
		},
		{ // path not registered
			handler:    fakeHandler{"na"},
			path:       "/posts",
			reqURL:     "https://example.com/posts/1",
			schemes:    []string{"https"},
			hostnames:  []string{"example.com"},
			statusCode: 404,
			params:     make(map[string]string),
		},
	}
	for _, test := range tests {
		req := httptest.NewRequest("GET", test.reqURL, nil)
		rw := httptest.NewRecorder()
		mockInterceptorFirst := &fakeInterceptor{"first"}
		mockInterceptorSecond := &fakeInterceptor{"second"}

		r, _ := New(
			WithSchemes(test.schemes...),
			WithHostnames(test.hostnames...),
			WithInterceptors(mockInterceptorFirst, mockInterceptorSecond),
		)

		_ = r.Get(test.path, test.handler)

		r.ServeHTTP(rw, req)

		t.Run("applies interceptors in correct order", func(t *testing.T) {
			if mockInterceptorFirst.name != "called" {
				t.Fatalf("interceptor must be executed on the serve")
			}
			if mockInterceptorSecond.name != "called" {
				t.Fatalf("interceptor must be executed on the serve")
			}
		})

		resp := rw.Result()
		t.Run("has correct status code", func(t *testing.T) {
			if resp.StatusCode != test.statusCode {
				t.Fatalf(
					"want status code %d, but got %d",
					test.statusCode,
					resp.StatusCode,
				)
			}
		})
	}
}

func TestMethodRegistrations(t *testing.T) {
	r, _ := New()
	getPath, getHandler := "/test-get-path", fakeHandler{"ok"}
	headPath, headHandler := "/test-head-path", fakeHandler{"ok"}
	postPath, postHandler := "/test-post-path", fakeHandler{"ok"}
	putPath, putHandler := "/test-put-path", fakeHandler{"ok"}
	patchPath, patchHandler := "/test-patch-path", fakeHandler{"ok"}
	deletePath, deleteHandler := "/test-delete-path", fakeHandler{"ok"}
	connectPath, connectHandler := "/test-connect-path", fakeHandler{"ok"}
	optionsPath, optionsHandler := "/test-options-path", fakeHandler{"ok"}
	tracePath, traceHandler := "/test-trace-path", fakeHandler{"ok"}

	tests := []struct {
		method  string
		path    string
		handler http.Handler
	}{
		{http.MethodGet, getPath, getHandler},
		{http.MethodHead, headPath, headHandler},
		{http.MethodPost, postPath, postHandler},
		{http.MethodPut, putPath, putHandler},
		{http.MethodPatch, patchPath, patchHandler},
		{http.MethodDelete, deletePath, deleteHandler},
		{http.MethodConnect, connectPath, connectHandler},
		{http.MethodOptions, optionsPath, optionsHandler},
		{http.MethodTrace, tracePath, traceHandler},
	}

	_ = r.Get(getPath, getHandler)
	_ = r.Head(headPath, headHandler)
	_ = r.Post(postPath, postHandler)
	_ = r.Put(putPath, putHandler)
	_ = r.Patch(patchPath, patchHandler)
	_ = r.Delete(deletePath, deleteHandler)
	_ = r.Connect(connectPath, connectHandler)
	_ = r.Options(optionsPath, optionsHandler)
	_ = r.Trace(tracePath, traceHandler)

	for _, test := range tests {
		t.Run("register handler for correct http method & path", func(t *testing.T) {
			path := []string{strings.Split(test.path, "/")[1]}
			h, ok := r.(*router).matcher.Find(test.method, path)
			if !ok || h.HTTPHandler != test.handler {
				t.Fatalf(
					"must register handler(%+v) for the path(%s) but got %+v",
					test.handler,
					test.path,
					h.HTTPHandler,
				)
			}
		})
	}

	t.Run("when HTTP handler registration fails", func(t *testing.T) {
		err := r.Get("/some", nil)
		if err == nil {
			t.Fatalf("registration of handler must fail handler is nil")
		}
	})
}

type fakeInterceptor struct {
	name string
}

func (m *fakeInterceptor) Middleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		m.name = "called"
		h.ServeHTTP(rw, req)
	})
}

type fakeHandler struct {
	bodyText string
}

func (h fakeHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	_, _ = rw.Write([]byte(h.bodyText))
}
