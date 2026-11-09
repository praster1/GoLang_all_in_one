package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
)

func gate(accepting *atomic.Bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !accepting.Load() {
			w.Header().Set("Retry-After", "1")
			http.Error(w, "draining", 503)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func main() {
	var accepting atomic.Bool
	accepting.Store(true)
	handler := gate(&accepting, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) }))
	for _, state := range []bool{true, false} {
		accepting.Store(state)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("POST", "/work", nil))
		fmt.Println("상태:", w.Code)
	}
}
