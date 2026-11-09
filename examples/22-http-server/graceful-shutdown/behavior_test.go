package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

type observedListener struct {
	net.Listener
	closed chan struct{}
	once   sync.Once
}

func (l *observedListener) Close() error {
	err := l.Listener.Close()
	l.once.Do(func() { close(l.closed) })
	return err
}
func TestShutdownWaitsForActiveRequest(t *testing.T) {
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener := &observedListener{Listener: raw, closed: make(chan struct{})}
	started := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseWork := func() { releaseOnce.Do(func() { close(release) }) }
	defer releaseWork()
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		if _, err := io.WriteString(w, "done"); err != nil {
			return
		}
	}), ReadHeaderTimeout: time.Second}
	defer func() {
		releaseWork()
		server.Close()
	}()
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	client := &http.Client{Timeout: 3 * time.Second}
	defer client.CloseIdleConnections()
	requestDone := make(chan error, 1)
	go func() {
		resp, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			requestDone <- err
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 16))
		if err == nil && string(body) != "done" {
			err = errors.New("incomplete response")
		}
		requestDone <- err
	}()
	select {
	case <-started:
	case err := <-requestDone:
		t.Fatalf("request ended early: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	shutdownDone := make(chan error, 1)
	go func() { shutdownDone <- server.Shutdown(ctx) }()
	<-listener.closed
	select {
	case err := <-shutdownDone:
		t.Fatalf("returned before active request ended: %v", err)
	default:
	}
	releaseWork()
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-shutdownDone; err != nil {
		t.Fatal(err)
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		t.Fatalf("Serve err=%v", err)
	}
}
