package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
		bad    bool
	}{
		{200, "ok\n", false}, {503, "ok\n", true}, {200, "ok\nextra", true}, {200, "", true},
	} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(tc.status); w.Write([]byte(tc.body)) }))
		client := server.Client()
		client.Timeout = time.Second
		err := probe(client, server.URL)
		server.Close()
		if (err != nil) != tc.bad {
			t.Fatalf("status=%d body=%q err=%v", tc.status, tc.body, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// smoke: ok
}
