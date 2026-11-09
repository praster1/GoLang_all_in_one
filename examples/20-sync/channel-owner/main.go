package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	inputs := [][]string{{"go", "test"}, {"go", "한글"}, {"test", "go"}}
	words := make(chan string)
	var group sync.WaitGroup
	for _, part := range inputs {
		group.Go(func() {
			for _, word := range part {
				words <- word
			}
		})
	}
	closed := make(chan struct{})
	go func() { group.Wait(); close(words); close(closed) }()
	counts := map[string]int{}
	for word := range words {
		counts[word]++
	}
	<-closed
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		fmt.Printf("%s: %d\n", key, counts[key])
	}
}
