package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServeHTTP(t *testing.T) {
	internalServerError := InternalServerError{}
	tests := []struct {
		handlerFunc http.HandlerFunc
		statusCode  int
	}{
		{
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				defer internalServerError.ServeHTTP(rw, req)
				io.WriteString(rw, "<html><body>Hello World!</body></html>")
			},
			statusCode: http.StatusOK,
		},
		{
			handlerFunc: func(rw http.ResponseWriter, req *http.Request) {
				defer internalServerError.ServeHTTP(rw, req)
				panic("ohh no!")
			},
			statusCode: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run("has correct status code", func(t *testing.T) {
			req := httptest.NewRequest("GET", "http://example.com/foo", nil)
			rw := httptest.NewRecorder()
			test.handlerFunc(rw, req)
			resp := rw.Result()

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
