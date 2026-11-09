package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
)

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /hello", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if _, err := io.WriteString(w, "안녕하세요\n"); err != nil {
			return
		}
	})
	return mux
}
func run(ctx context.Context, args []string, out, diagnostics io.Writer) error {
	flags := flag.NewFlagSet("managed-server", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	serve := flags.Bool("serve", false, "직접 HTTP 요청을 받습니다")
	address := flags.String("addr", "127.0.0.1:8080", "loopback 주소")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("위치 인자는 받지 않습니다")
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil {
		return errors.New("주소는 호스트:포트 형식이어야 합니다")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("학습 서버는 loopback 주소만 사용합니다")
	}
	if !*serve {
		w := httptest.NewRecorder()
		handler().ServeHTTP(w, httptest.NewRequest("GET", "/hello", nil))
		_, err := fmt.Fprintf(out, "상태: %d\n%s", w.Code, w.Body.String())
		return err
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return err
	}
	logger := slog.New(slog.NewJSONHandler(diagnostics, nil))
	server := &http.Server{Handler: handler(), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	logger.Info("listening", "address", listener.Addr().String())
	err = service.Serve(ctx, server, listener, 3*time.Second)
	if err != nil {
		logger.Error("shutdown failed", "error", err)
		return err
	}
	logger.Info("stopped")
	return nil
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	err := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
