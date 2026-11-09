package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/praster1/GoLang_all_in_one/examples/31-concurrent-batch/batch"
)

func main() {
	calls := 0
	process := func(context.Context, batch.Job) (string, error) { calls++; return "ok", nil }
	results, err := batch.Run(context.Background(), []string{"go"}, batch.Config{Workers: 0, Attempts: 1}, process)
	fmt.Println("설정 거부:", results == nil && errors.Is(err, batch.ErrConfig))
	results, err = batch.Run(context.Background(), []string{strings.Repeat("x", batch.MaxItemBytes+1)}, batch.Config{Workers: 1, Attempts: 1}, process)
	fmt.Println("입력 거부:", results == nil && errors.Is(err, batch.ErrInput))
	fmt.Println("실행된 작업:", calls)
}
