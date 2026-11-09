package main

import (
	"context"
	"fmt"
)

func main() {
	ctx := context.Background()
	_, hasDeadline := ctx.Deadline()
	fmt.Println("마감 있음:", hasDeadline)
	fmt.Println("취소됨:", ctx.Err() != nil)
	fmt.Println("Done 없음:", ctx.Done() == nil)
}
