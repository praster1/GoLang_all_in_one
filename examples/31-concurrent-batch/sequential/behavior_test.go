package main

import (
	"context"
	"errors"
	"testing"
)

func TestSequentialBoundaries(t *testing.T) {
	if got, err := sequential(context.Background(), nil); err != nil || got == nil || len(got) != 0 {
		t.Fatal(got, err)
	}
	for _, input := range [][]string{{" "}, make([]string, 1001)} {
		if got, err := sequential(context.Background(), input); got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if got, err := sequential(ctx, []string{"go"}); got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
}
