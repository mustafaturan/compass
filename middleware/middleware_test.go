package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNext(t *testing.T) {
	mfn := Func(func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Header().Set("X-Middleware", "compass")
			h.ServeHTTP(rw, req)
		})
	})

	hfn := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("pong"))
	})

	rw := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "http://example.com/ping", nil)
	mfn.Next(hfn).(http.Handler).ServeHTTP(rw, req)

	resp := rw.Result()
	t.Run("verify middleware effect with http header", func(t *testing.T) {
		if resp.Header.Get("X-Middleware") != "compass" {
			t.Fatalf("middleware Next() should call Next() with effects")
		}
		if body, err := ioutil.ReadAll(resp.Body); string(body) != "pong" || err != nil {
			t.Fatalf("middleware Next() should call Next() with effects %+v", err)
		}
	})
}
