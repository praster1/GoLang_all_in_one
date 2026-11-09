package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestTraceID(t *testing.T) {
	base := context.Background()
	if id, ok := traceID(base); id != "" || ok {
		t.Fatal(id, ok)
	}
	for _, input := range []string{"", " "} {
		ctx, err := withTraceID(base, input)
		if err == nil || ctx != nil {
			t.Fatal(ctx, err)
		}
	}
	parent, err := withTraceID(base, " parent ")
	if err != nil {
		t.Fatal(err)
	}
	child, err := withTraceID(parent, "child")
	if err != nil {
		t.Fatal(err)
	}
	if id, _ := traceID(parent); id != "parent" {
		t.Fatal(id)
	}
	if id, _ := traceID(child); id != "child" {
		t.Fatal(id)
	}
	type unrelatedKey struct{}
	other := context.WithValue(parent, unrelatedKey{}, "other")
	if id, _ := traceID(other); id != "parent" {
		t.Fatal(id)
	}
	expired, cancel := context.WithDeadline(base, time.Unix(0, 0))
	defer cancel()
	child, err = withTraceID(expired, "id")
	if err != nil {
		t.Fatal(err)
	}
	if !errors.Is(child.Err(), context.DeadlineExceeded) {
		t.Fatal(child.Err())
	}
	if deadline, ok := child.Deadline(); !ok || !deadline.Equal(time.Unix(0, 0)) {
		t.Fatal(deadline, ok)
	}
}

func Example_main() {
	main()
	// Output:
	// 추적 ID: trace-001 true
	// 취소 전달: true
	// 값 유지: trace-001 true
}
