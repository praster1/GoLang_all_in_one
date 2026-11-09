package main

import (
	"crypto/sha256"
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
)

func bearerAuth(expected string, next http.Handler) (http.Handler, error) {
	if len(expected) < 16 || strings.ContainsAny(expected, " \t\r\n") {
		return nil, errors.New("설정 토큰은 공백 없는 16바이트 이상이어야 합니다")
	}
	want := sha256.Sum256([]byte(expected))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.Header.Values("Authorization")
		if len(values) == 1 && strings.HasPrefix(values[0], "Bearer ") {
			got := sha256.Sum256([]byte(strings.TrimPrefix(values[0], "Bearer ")))
			if subtle.ConstantTimeCompare(got[:], want[:]) == 1 {
				next.ServeHTTP(w, r)
				return
			}
		}
		w.Header().Set("WWW-Authenticate", "Bearer")
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}), nil
}
func run() error {
	token := "classroom-demo-token"
	handler, err := bearerAuth(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	if err != nil {
		return err
	}
	for _, credential := range []string{"", "wrong-token", token} {
		request := httptest.NewRequest("POST", "/private", nil)
		if credential != "" {
			request.Header.Set("Authorization", "Bearer "+credential)
		}
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		fmt.Println("상태:", recorder.Code)
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
