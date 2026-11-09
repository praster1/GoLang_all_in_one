package service

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

type watchedListener struct {
	net.Listener
	closed chan struct{}
	once   sync.Once
}

func (l *watchedListener) Close() error {
	err := l.Listener.Close()
	l.once.Do(func() { close(l.closed) })
	return err
}
func wait(t *testing.T, ch <-chan struct{}) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("operation did not finish")
	}
}
func TestServeDrainsActiveRequest(t *testing.T) {
	raw, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	listener := &watchedListener{Listener: raw, closed: make(chan struct{})}
	entered := make(chan struct{})
	release := make(chan struct{})
	var releaseOnce sync.Once
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-release; w.WriteHeader(204) })}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stopped := make(chan error, 1)
	go func() { stopped <- Serve(ctx, server, listener, time.Second) }()
	t.Cleanup(func() { releaseOnce.Do(func() { close(release) }); _ = server.Close() })
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	requestDone := make(chan error, 1)
	go func() {
		response, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_, readErr := io.Copy(io.Discard, response.Body)
			err = errors.Join(readErr, response.Body.Close())
			if response.StatusCode != 204 {
				err = errors.New("wrong response")
			}
		}
		requestDone <- err
	}()
	wait(t, entered)
	cancel()
	wait(t, listener.closed)
	select {
	case err := <-stopped:
		t.Fatalf("returned before handler: %v", err)
	default:
	}
	releaseOnce.Do(func() { close(release) })
	if err := <-requestDone; err != nil {
		t.Fatal(err)
	}
	if err := <-stopped; err != nil {
		t.Fatal(err)
	}
}
func TestServeTimeoutCancelsActiveRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	done := make(chan struct{})
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(entered); <-r.Context().Done(); close(done) })}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer server.Close()
	stopped := make(chan error, 1)
	go func() { stopped <- Serve(ctx, server, listener, 10*time.Millisecond) }()
	client := &http.Client{Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	requestDone := make(chan struct{})
	go func() {
		defer close(requestDone)
		resp, err := client.Get("http://" + listener.Addr().String())
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	wait(t, entered)
	cancel()
	if err := <-stopped; !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error=%v", err)
	}
	wait(t, done)
	wait(t, requestDone)
}
func TestServeStartupCanceled(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Serve(ctx, &http.Server{}, listener, time.Second); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
	if _, err := listener.Accept(); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("listener error=%v", err)
	}
}
