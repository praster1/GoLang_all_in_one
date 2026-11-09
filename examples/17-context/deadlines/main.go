package main

import (
	"context"
	"errors"
	"fmt"
	"time"
)

func main() {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	<-ctx.Done()
	deadline, exists := ctx.Deadline()
	fmt.Println("마감 있음:", exists)
	fmt.Println("마감:", deadline.UTC().Format(time.RFC3339))
	fmt.Println("마감 초과:", errors.Is(ctx.Err(), context.DeadlineExceeded))

	immediate, stop := context.WithTimeout(context.Background(), 0)
	defer stop()
	<-immediate.Done()
	fmt.Println("0초 제한:", errors.Is(immediate.Err(), context.DeadlineExceeded))
}
