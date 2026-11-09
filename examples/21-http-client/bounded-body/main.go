package main

import (
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
	if limit < 0 || limit > 1<<20 {
		return nil, errors.New("한도는 0~1048576바이트")
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

func download(client *http.Client, address string, limit int64) ([]byte, error) {
	resp, err := client.Get(address)
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
		if _, err := io.WriteString(w, "ABCDE"); err != nil {
			return
		}
	}))
	defer server.Close()
	client := server.Client()
	client.Timeout = 2 * time.Second
	defer client.CloseIdleConnections()
	_, err := download(client, server.URL, 4)
	if !errors.Is(err, ErrTooLarge) {
		return fmt.Errorf("예상과 다른 결과: %v", err)
	}
	fmt.Println("4바이트 제한: 초과")
	body, err := download(client, server.URL, 5)
	if err != nil {
		return err
	}
	fmt.Println("5바이트 제한:", string(body))
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
