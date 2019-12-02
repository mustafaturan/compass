package matcher

import (
	"net/http"
	"testing"

	chandler "github.com/mustafaturan/compass/handler"
)

func TestNew(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{http.MethodGet, true},
		{http.MethodHead, true},
		{http.MethodPost, true},
		{http.MethodPut, true},
		{http.MethodPatch, true},
		{http.MethodDelete, true},
		{http.MethodConnect, true},
		{http.MethodOptions, true},
		{http.MethodTrace, true},
		{"na", false},
	}
	m := New()
	t.Run("inits with right keys", func(t *testing.T) {
		for _, test := range tests {
			if _, got := m.nodes[test.method]; got != test.want {
				t.Fatalf("New() expected to register %s node as %v", test.method, test.want)
			}
		}
	})
}

func TestRegister(t *testing.T) {
	routes := []struct {
		method string
		path   string
	}{
		{
			method: http.MethodGet,
			path:   "/posts",
		},
		{
			method: http.MethodGet,
			path:   "/posts/:id",
		},
		{
			method: http.MethodGet,
			path:   "/posts/:id/comments",
		},
		{
			method: http.MethodGet,
			path:   "/posts/:id/comments/commentID",
		},
		{
			method: http.MethodGet,
			path:   "/posts/:id/reviews",
		},
		{
			method: http.MethodGet,
			path:   "/posts/:id/reviews/:reviewID",
		},
	}

	m := New()

	t.Run("first time registrations should not return error", func(t *testing.T) {
		for _, route := range routes {
			handler, _ := chandler.New(route.path, testHTTPHandler{})
			if err := m.Register(route.method, handler); err != nil {
				t.Fatalf(
					"Register(%s, %+v) SHOULD NOT return err %s",
					route.method,
					handler,
					err,
				)
			}
		}
	})

	t.Run("registrations for the same route should return error", func(t *testing.T) {
		for _, route := range routes {
			handler, _ := chandler.New(route.path, testHTTPHandler{})
			if err := m.Register(route.method, handler); err == nil {
				t.Fatalf(
					"Register(%s, %+v) SHOULD return err",
					route.method,
					handler,
				)
			}
		}
	})

	t.Run("registration of the same path should return error", func(t *testing.T) {
		handler, _ := chandler.New("/posts/:id/reviews/9", testHTTPHandler{})
		if err := m.Register(http.MethodGet, handler); err == nil {
			t.Fatalf(
				"Register(%s, %+v) SHOULD return err",
				http.MethodGet,
				handler,
			)
		}
	})
}

func TestFind(t *testing.T) {
	routes := []struct {
		path    string
		method  string
		handler *chandler.Handler
	}{
		{
			path:   "/",
			method: http.MethodGet,
		},
		{
			path:   "/posts",
			method: http.MethodGet,
		},
		{
			path:   "/posts/-1",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/comments",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/comments/:commentID",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/reviews",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/reviews/:reviewID",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/reviews/101/likers",
			method: http.MethodGet,
		},
		{
			path:   "/posts/:id/reviews/:reviewID/likers/:liker",
			method: http.MethodGet,
		},
	}

	m := New()
	for i, route := range routes {
		routes[i].handler, _ = chandler.New(route.path, testHTTPHandler{})

		if err := m.Register(route.method, routes[i].handler); err != nil {
			panic(err)
		}
	}

	tests := []struct {
		method      string
		path        []string
		wantHandler *chandler.Handler
		wantResult  bool
	}{
		// Available routes
		{http.MethodGet, []string{}, routes[0].handler, true},
		{http.MethodGet, []string{"posts"}, routes[1].handler, true},
		{http.MethodGet, []string{"posts", "99"}, routes[3].handler, true},
		{http.MethodGet, []string{"posts", "99", "comments"}, routes[4].handler, true},
		{http.MethodGet, []string{"posts", "99", "comments", "56"}, routes[5].handler, true},
		{http.MethodGet, []string{"posts", "99", "reviews"}, routes[6].handler, true},
		{http.MethodGet, []string{"posts", "99", "reviews", "56"}, routes[7].handler, true},
		{http.MethodGet, []string{"posts", "99", "reviews", "101", "likers"}, routes[8].handler, true},
		{http.MethodGet, []string{"posts", "99", "reviews", "56", "likers", "33"}, routes[9].handler, true},
		{http.MethodGet, []string{"posts", "-1"}, routes[2].handler, true},

		// Non-existed routes
		{http.MethodGet, []string{"comments"}, nil, false},
		{http.MethodGet, []string{"comments", "76"}, nil, false},
		{http.MethodGet, []string{"reviews"}, nil, false},
		{http.MethodGet, []string{"reviews", "33"}, nil, false},
		{http.MethodGet, []string{"posts", "99", "reviews", "56", "likers"}, nil, false},
	}

	for _, test := range tests {
		gotHandler, gotResult := m.Find(test.method, test.path)

		t.Run("correct handler match", func(t *testing.T) {
			if gotHandler != test.wantHandler {
				t.Fatalf(
					"Find(%s, %v) should result with %+v but got %+v",
					test.method,
					test.path,
					test.wantHandler,
					gotHandler,
				)
			}
		})

		t.Run("route matching estimations", func(t *testing.T) {
			if gotResult != test.wantResult {
				t.Fatalf(
					"Find(%s, %v) should result with %v but got %v",
					test.method,
					test.path,
					test.wantResult,
					gotResult,
				)
			}
		})
	}
}

type testHTTPHandler struct{}

func (h testHTTPHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
}
