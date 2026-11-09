package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"
)

func search(ctx context.Context, client *http.Client, address, term string) (string, error) {
	u, err := url.Parse(address)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set("q", term)
	u.RawQuery = query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept-Language", "ko-KR")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("상태: %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1025))
	if err != nil {
		return "", err
	}
	if len(body) > 1024 {
		return "", fmt.Errorf("본문 한도 초과")
	}
	return string(body), nil
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintf(w, "검색어: %s\n언어: %s\n", r.URL.Query().Get("q"), r.Header.Get("Accept-Language")); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	text, err := search(context.Background(), client, server.URL, "Go & HTTP")
	if err != nil {
		return err
	}
	fmt.Print(text)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
