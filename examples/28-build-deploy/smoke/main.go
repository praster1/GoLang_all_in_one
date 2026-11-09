package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"time"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "ok")
	})
	return mux
}

func probe(client *http.Client, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health status: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4))
	if err != nil {
		return err
	}
	if string(body) != "ok\n" {
		return fmt.Errorf("unexpected health body")
	}
	return nil
}

func main() {
	server := httptest.NewServer(handler())
	defer server.Close()
	client := server.Client()
	client.Timeout = time.Second
	if err := probe(client, server.URL+"/healthz"); err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("smoke: ok")
}
