package main

import (
	"context"
	"fmt"
	"io"
	"os"
)

func Example_run() {
	if err := run(context.Background(), nil, os.Stdout, io.Discard); err != nil {
		fmt.Println(err)
	}
	// Output:
	// 상태: 200
	// 안녕하세요
}
