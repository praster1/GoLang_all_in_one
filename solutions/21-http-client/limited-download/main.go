package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"time"
)

var ErrTooLarge = errors.New("본문 한도 초과")

func readAtMost(reader io.Reader, limit int64) ([]byte, error) {
	if limit < 0 || limit > 1024 {
		return nil, errors.New("한도는 0~1024바이트")
	}
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, ErrTooLarge
	}
	return body, nil
}

func download(ctx context.Context, client *http.Client, address string, limit int64) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("상태: %d", resp.StatusCode)
	}
	return readAtMost(resp.Body, limit)
}

func run() error {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "hello"); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	_, err := download(context.Background(), client, server.URL, 4)
	if !errors.Is(err, ErrTooLarge) {
		return fmt.Errorf("예상과 다른 결과: %v", err)
	}
	fmt.Println("한도 초과: true")
	body, err := download(context.Background(), client, server.URL, 5)
	if err != nil {
		return err
	}
	fmt.Println("본문:", string(body))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
