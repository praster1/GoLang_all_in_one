package main

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

type Inventory struct {
	mu    sync.Mutex
	stock int
}

func newInventory(stock int) (*Inventory, error) {
	if stock < 0 || stock > 10000 {
		return nil, errors.New("재고는 0~10000이어야 합니다")
	}
	return &Inventory{stock: stock}, nil
}
func (i *Inventory) Reserve(quantity int) bool {
	if quantity < 1 || quantity > 100 {
		return false
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.stock < quantity {
		return false
	}
	i.stock -= quantity
	return true
}
func (i *Inventory) Remaining() int { i.mu.Lock(); defer i.mu.Unlock(); return i.stock }
func main() {
	inventory, err := newInventory(10)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	results := make([]bool, 20)
	var group sync.WaitGroup
	for index := range results {
		group.Go(func() { results[index] = inventory.Reserve(1) })
	}
	group.Wait()
	successes := 0
	for _, ok := range results {
		if ok {
			successes++
		}
	}
	fmt.Println("예약 성공:", successes)
	fmt.Println("남은 재고:", inventory.Remaining())
}
