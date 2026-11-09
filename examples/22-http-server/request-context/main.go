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

func waitForWork(ctx context.Context) error {
	<-ctx.Done()
	return ctx.Err()
}

func run() error {
	started := make(chan struct{})
	finished := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		err := waitForWork(r.Context())
		finished <- err
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		return err
	}
	result := make(chan error, 1)
	go func() {
		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
		result <- err
	}()
	select {
	case <-started:
		cancel()
	case err := <-result:
		return fmt.Errorf("시작 전 종료: %v", err)
	}
	if err := <-result; !errors.Is(err, context.Canceled) {
		return fmt.Errorf("클라이언트 결과: %v", err)
	}
	if err := <-finished; !errors.Is(err, context.Canceled) {
		return fmt.Errorf("서버 결과: %v", err)
	}
	fmt.Println("클라이언트 취소: true")
	fmt.Println("서버 작업 취소: true")
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
