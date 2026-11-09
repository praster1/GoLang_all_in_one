package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	reason := errors.New("사용자가 작업 중단")
	cancel(reason)
	<-ctx.Done()
	fmt.Println("상태:", ctx.Err())
	fmt.Println("원인:", context.Cause(ctx))
	cancel(errors.New("나중의 원인"))
	fmt.Println("첫 원인 유지:", errors.Is(context.Cause(ctx), reason))
}
