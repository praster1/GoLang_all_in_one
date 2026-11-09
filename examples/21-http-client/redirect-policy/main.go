package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func inspect(client *http.Client, address string) (int, string, error) {
	resp, err := client.Get(address)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	if _, err := io.Copy(io.Discard, io.LimitReader(resp.Body, 1024)); err != nil {
		return 0, "", err
	}
	return resp.StatusCode, resp.Header.Get("Location"), nil
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	mux.HandleFunc("/final", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "완료"); err != nil {
			return
		}
	})
	server := httptest.NewTLSServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	code, _, err := inspect(client, server.URL+"/start")
	if err != nil {
		return err
	}
	fmt.Println("기본 정책:", code)
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	code, location, err := inspect(client, server.URL+"/start")
	if err != nil {
		return err
	}
	fmt.Println("중지 정책:", code)
	fmt.Println("위치:", location)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
