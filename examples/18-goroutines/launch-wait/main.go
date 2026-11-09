package main

import (
	"fmt"
	"sync"
)

func main() {
	var result string
	var group sync.WaitGroup
	group.Add(1)
	go func() {
		defer group.Done()
		result = "고루틴 작업 완료"
	}()
	group.Wait()
	fmt.Println(result)
	fmt.Println("main 종료")
}
