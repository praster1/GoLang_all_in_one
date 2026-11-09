package main

import (
	"context"
	"fmt"
	"sync"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{})
	stopped := false
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		close(started)
		<-ctx.Done()
		stopped = true
	}()
	<-started
	cancel()
	group.Wait()
	fmt.Println("시작 확인:", true)
	fmt.Println("종료 확인:", stopped)
}
