package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

func request(ctx context.Context, client *http.Client, address string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP 호출: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func run() error {
	started := make(chan struct{})
	stopped := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(stopped)
		close(started)
		<-r.Context().Done()
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		result <- request(ctx, client, server.URL)
	}()
	select {
	case <-started:
		cancel()
	case err := <-result:
		return fmt.Errorf("취소 전 요청 종료: %v", err)
	}
	err := <-result
	<-stopped
	if !errors.Is(err, context.Canceled) {
		return fmt.Errorf("취소 결과: %v", err)
	}
	fmt.Println("요청 취소 확인: true")
	fmt.Println("서버 작업 종료: true")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
