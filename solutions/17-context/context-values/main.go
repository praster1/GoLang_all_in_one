package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

type traceIDKey struct{}

func withTraceID(ctx context.Context, id string) (context.Context, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, errors.New("추적 ID는 비어 있을 수 없습니다")
	}
	return context.WithValue(ctx, traceIDKey{}, id), nil
}

func traceID(ctx context.Context) (string, bool) {
	id, exists := ctx.Value(traceIDKey{}).(string)
	return id, exists
}

func run() error {
	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	child, err := withTraceID(parent, " trace-001 ")
	if err != nil {
		return err
	}
	id, exists := traceID(child)
	fmt.Println("추적 ID:", id, exists)
	cancel()
	<-child.Done()
	fmt.Println("취소 전달:", errors.Is(child.Err(), context.Canceled))
	id, exists = traceID(child)
	fmt.Println("값 유지:", id, exists)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
