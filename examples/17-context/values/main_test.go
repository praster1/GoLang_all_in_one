package main

import (
	"context"
	"testing"
)

func TestRequestID(t *testing.T) {
	base := context.Background()
	if id, ok := requestID(base); id != "" || ok {
		t.Fatal(id, ok)
	}
	parent := withRequestID(base, "parent")
	child := withRequestID(parent, "child")
	if id, _ := requestID(parent); id != "parent" {
		t.Fatal(id)
	}
	if id, _ := requestID(child); id != "child" {
		t.Fatal(id)
	}
	if id, ok := requestID(withRequestID(base, "")); id != "" || !ok {
		t.Fatal("empty value exists", id, ok)
	}
}

func Example_main() {
	main()
	// Output:
	// 요청 ID: req-001
	// 값 있음: true
	// 취소 후 값: req-001 true
}
