package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

type Slots struct {
	limit  int64
	active atomic.Int64
}

func newSlots(limit int64) (*Slots, error) {
	if limit < 1 || limit > 32 {
		return nil, errors.New("슬롯 수는 1~32여야 합니다")
	}
	return &Slots{limit: limit}, nil
}
func (s *Slots) TryAcquire() (func(), bool) {
	for {
		old := s.active.Load()
		if old >= s.limit {
			return nil, false
		}
		if s.active.CompareAndSwap(old, old+1) {
			var once sync.Once
			release := func() { once.Do(func() { s.active.Add(-1) }) }
			return release, true
		}
	}
}
func main() {
	slots, err := newSlots(2)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	first, ok1 := slots.TryAcquire()
	second, ok2 := slots.TryAcquire()
	_, ok3 := slots.TryAcquire()
	fmt.Println("처음 세 요청:", ok1, ok2, ok3)
	first()
	first()
	third, ok4 := slots.TryAcquire()
	fmt.Println("반납 후 획득:", ok4)
	second()
	third()
	fmt.Println("남은 사용량:", slots.active.Load())
}
