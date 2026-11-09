package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"time"
)

func requestStatus(ctx context.Context, client *http.Client, address string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return 0, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("상태 조회: %w", err)
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func run() error {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := requestStatus(ctx, client, server.URL)
	if !errors.Is(err, context.Canceled) {
		return fmt.Errorf("취소 결과: %v", err)
	}
	fmt.Println("취소 확인: true")
	fmt.Println("서버 요청 수:", calls.Load())
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
