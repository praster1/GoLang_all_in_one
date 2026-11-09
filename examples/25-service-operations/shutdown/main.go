package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
)

func run() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(context.Background())
	server := &http.Server{ReadHeaderTimeout: time.Second, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})}
	result := make(chan error, 1)
	go func() { result <- service.Serve(ctx, server, listener, time.Second) }()
	client := &http.Client{Timeout: time.Second}
	defer client.CloseIdleConnections()
	response, requestErr := client.Get("http://" + listener.Addr().String())
	var readErr, closeErr error
	if requestErr == nil {
		_, readErr = io.Copy(io.Discard, response.Body)
		closeErr = response.Body.Close()
		fmt.Println("요청 상태:", response.StatusCode)
	}
	cancel()
	stopErr := <-result
	if err := errors.Join(requestErr, readErr, closeErr, stopErr); err != nil {
		return err
	}
	fmt.Println("종료 완료:", true)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
