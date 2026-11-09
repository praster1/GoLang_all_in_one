package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
)

func TestDemo(t *testing.T) {
	var out bytes.Buffer
	if err := run(context.Background(), nil, &out, io.Discard); err != nil {
		t.Fatal(err)
	}
	if out.String() != "상태: 200\n안녕하세요\n" {
		t.Fatalf("output=%q", out.String())
	}
}
func TestArguments(t *testing.T) {
	for _, args := range [][]string{{"-addr=0.0.0.0:8080"}, {"-addr=bad"}, {"extra"}, {"-unknown"}} {
		if err := run(context.Background(), args, io.Discard, io.Discard); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}
func TestCanceledBeforeListen(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, []string{"-serve", "-addr=127.0.0.1:0"}, io.Discard, io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
