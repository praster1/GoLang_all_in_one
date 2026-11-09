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

type StatusError struct {
	Code int
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("HTTP 상태 %d", e.Code)
}

func fetch(ctx context.Context, client *http.Client, address string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("전송: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", &StatusError{Code: resp.StatusCode}
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 129))
	if err != nil {
		return "", err
	}
	if len(body) > 128 {
		return "", errors.New("본문 한도 초과")
	}
	return string(body), nil
}

func run() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.WriteString(w, "준비 완료"); err != nil {
			return
		}
	})
	mux.HandleFunc("/busy", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "busy", http.StatusServiceUnavailable)
	})
	server := httptest.NewServer(mux)
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	for _, path := range []string{"/missing", "/busy", "/ok"} {
		text, err := fetch(context.Background(), client, server.URL+path)
		var status *StatusError
		if errors.As(err, &status) {
			fmt.Println("실패 상태:", status.Code)
			continue
		}
		if err != nil {
			return err
		}
		fmt.Println("본문:", text)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
