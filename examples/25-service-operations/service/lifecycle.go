// Package service connects a server lifetime to a cancellation signal.
package service

import (
	"context"
	"errors"
	"net"
	"net/http"
	"time"
)

// Serve takes ownership of listener. A canceled ctx requests graceful shutdown.
// Handlers must honor request cancellation if a forced close becomes necessary.
func Serve(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration) error {
	if timeout <= 0 {
		return errors.Join(errors.New("종료 제한 시간은 양수여야 합니다"), listener.Close())
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, listener.Close())
	}
	result := make(chan error, 1)
	go func() { result <- server.Serve(listener) }()
	select {
	case err := <-result:
		return errors.Join(err, server.Close())
	case <-ctx.Done():
		drain, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		shutdownErr := server.Shutdown(drain)
		var closeErr error
		if shutdownErr != nil {
			closeErr = server.Close()
		}
		serveErr := <-result
		if errors.Is(serveErr, http.ErrServerClosed) {
			serveErr = nil
		}
		return errors.Join(shutdownErr, closeErr, serveErr)
	}
}
