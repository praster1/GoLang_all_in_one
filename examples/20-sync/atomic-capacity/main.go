package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type Quota struct {
	used  atomic.Int64
	limit int64
}

func (q *Quota) Reserve(amount int64) bool {
	if amount <= 0 || amount > q.limit {
		return false
	}
	for {
		old := q.used.Load()
		if old > q.limit-amount {
			return false
		}
		if q.used.CompareAndSwap(old, old+amount) {
			return true
		}
	}
}
func main() {
	quota := Quota{limit: 10}
	successes := make([]bool, 20)
	var group sync.WaitGroup
	for index := range successes {
		group.Go(func() { successes[index] = quota.Reserve(1) })
	}
	group.Wait()
	count := 0
	for _, success := range successes {
		if success {
			count++
		}
	}
	fmt.Println("예약 성공:", count)
	fmt.Println("사용량:", quota.used.Load())
}
