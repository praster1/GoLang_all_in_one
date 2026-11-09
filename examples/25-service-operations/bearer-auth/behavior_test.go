package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuth(t *testing.T) {
	token := "classroom-test-token"
	calls := 0
	handler, err := bearerAuth(token, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls++; w.WriteHeader(204) }))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		values []string
		want   int
	}{
		{nil, 401}, {[]string{"Bearer wrong"}, 401}, {[]string{"Basic " + token}, 401}, {[]string{"Bearer " + token}, 204}, {[]string{"Bearer " + token, "Bearer " + token}, 401}, {[]string{"Bearer " + token + " "}, 401},
	} {
		r := httptest.NewRequest("GET", "/private", nil)
		for _, v := range tc.values {
			r.Header.Add("Authorization", v)
		}
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		if w.Code != tc.want {
			t.Fatalf("status=%d want=%d", w.Code, tc.want)
		}
		if tc.want == 401 && w.Header().Get("WWW-Authenticate") != "Bearer" {
			t.Fatal("missing challenge")
		}
	}
	if calls != 1 {
		t.Fatalf("called %d times", calls)
	}
	for _, value := range []string{"", "short", "long-but-has space"} {
		if _, err := bearerAuth(value, http.NotFoundHandler()); err == nil {
			t.Fatal("accepted invalid config")
		}
	}
}
