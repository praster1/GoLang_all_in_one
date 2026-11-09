package main

import (
	"context"
	"fmt"
)

type requestIDKey struct{}

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

func requestID(ctx context.Context) (string, bool) {
	id, exists := ctx.Value(requestIDKey{}).(string)
	return id, exists
}

func main() {
	parent := withRequestID(context.Background(), "req-001")
	child, cancel := context.WithCancel(parent)
	defer cancel()
	id, exists := requestID(child)
	fmt.Println("요청 ID:", id)
	fmt.Println("값 있음:", exists)
	cancel()
	<-child.Done()
	id, exists = requestID(child)
	fmt.Println("취소 후 값:", id, exists)
}
