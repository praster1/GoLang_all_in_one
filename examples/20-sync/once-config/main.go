package main

import (
	"fmt"
	"sync"
)

type Config struct{ Workers int }

func main() {
	calls := 0
	load := sync.OnceValues(func() (Config, error) { calls++; return Config{Workers: 4}, nil })
	results := make([]Config, 3)
	failures := make([]error, 3)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { results[index], failures[index] = load() })
	}
	group.Wait()
	fmt.Println("설정 읽기 횟수:", calls)
	fmt.Println("작업자 설정:", results[0].Workers, results[1].Workers, results[2].Workers)
	fmt.Println("오류 없음:", failures[0] == nil && failures[1] == nil && failures[2] == nil)
}
