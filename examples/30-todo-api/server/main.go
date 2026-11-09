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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/praster1/GoLang_all_in_one/examples/25-service-operations/service"
	"github.com/praster1/GoLang_all_in_one/examples/30-todo-api/todo"
)

func requestDeadline(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func run(ctx context.Context, args []string, diagnostics io.Writer) (err error) {
	flags := flag.NewFlagSet("todo-api", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	address := flags.String("addr", "127.0.0.1:8080", "loopback 호스트:포트")
	store := flags.String("store", "memory", "memory 또는 sqlite")
	path := flags.String("db", "", "sqlite를 선택했을 때 사용할 파일 경로")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("위치 인자는 받지 않습니다")
	}
	host, _, err := net.SplitHostPort(*address)
	if err != nil {
		return errors.New("올바른 호스트:포트가 필요합니다")
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() {
		return errors.New("학습 서버는 loopback IP만 사용합니다")
	}
	if *store != "memory" && *store != "sqlite" {
		return errors.New("store는 memory 또는 sqlite여야 합니다")
	}
	if *store == "sqlite" && *path == "" {
		return errors.New("sqlite에는 -db 경로가 필요합니다")
	}
	if *store == "memory" && *path != "" {
		return errors.New("memory 모드에는 -db를 지정하지 않습니다")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var repository todo.Repository
	if *store == "sqlite" {
		startup, cancel := context.WithTimeout(ctx, 3*time.Second)
		db, openErr := todo.OpenSQLite(startup, *path)
		cancel()
		if openErr != nil {
			return openErr
		}
		defer func() { err = errors.Join(err, db.Close()) }()
		repository = db
	} else {
		repository = todo.NewMemory()
	}
	logger := slog.New(slog.NewJSONHandler(diagnostics, nil)).With("service", "todo-api")
	handler, err := todo.NewHandler(repository, logger)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", *address)
	if err != nil {
		return err
	}
	server := &http.Server{Handler: requestDeadline(handler), ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second, WriteTimeout: 5 * time.Second, IdleTimeout: 30 * time.Second}
	logger.Info("listening", "address", listener.Addr().String(), "store", *store)
	if err := service.Serve(ctx, server, listener, 3*time.Second); err != nil {
		return err
	}
	logger.Info("stopped")
	return nil
}
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := run(ctx, os.Args[1:], os.Stderr)
	stop()
	if err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
