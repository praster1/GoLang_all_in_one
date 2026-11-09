package main

import (
	"context"
	"errors"
	"fmt"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	fmt.Println("취소 전:", ctx.Err() == nil)
	cancel()
	cancel()
	<-ctx.Done()
	fmt.Println("취소 확인:", errors.Is(ctx.Err(), context.Canceled))
	fmt.Println("부모 취소:", context.Background().Err() != nil)
}
