package interceptor

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddleware(t *testing.T) {
	mfn := Func(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("X-Middleware", "compass")
			h.ServeHTTP(rw, req)
		})
	})

	hfn := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		_, _ = rw.Write([]byte("pong"))
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/ping", nil)
	mfn.Middleware(hfn).(http.Handler).ServeHTTP(rw, req)

	resp := rw.Result()
	t.Run("verify interceptor effect with http header", func(t *testing.T) {
		if resp.Header.Get("X-Middleware") != "compass" {
			t.Fatalf("interceptor Middleware() should call Middleware() with effects")
		}
		if body, err := ioutil.ReadAll(resp.Body); string(body) != "pong" || err != nil {
			t.Fatalf("interceptor Middleware() should call Middleware() with effects %+v", err)
		}
	})
}
