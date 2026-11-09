package main

import (
	"errors"
	"fmt"
	"sync"
)

type Ledger struct {
	mu       sync.Mutex
	balances map[string]int
}

func newLedger() *Ledger { return &Ledger{balances: map[string]int{"A": 100, "B": 100}} }
func (l *Ledger) Transfer(from, to string, amount int) error {
	if amount <= 0 || amount > 100 {
		return errors.New("금액은 1~100이어야 합니다")
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	source, exists := l.balances[from]
	if !exists {
		return errors.New("출금 계정이 없습니다")
	}
	destination, exists := l.balances[to]
	if !exists {
		return errors.New("입금 계정이 없습니다")
	}
	if from == to {
		return errors.New("같은 계정으로 옮길 수 없습니다")
	}
	if source < amount {
		return errors.New("잔액이 부족합니다")
	}
	l.balances[from] = source - amount
	l.balances[to] = destination + amount
	return nil
}
func (l *Ledger) Snapshot() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	copy := make(map[string]int, len(l.balances))
	for key, value := range l.balances {
		copy[key] = value
	}
	return copy
}
func main() {
	ledger := newLedger()
	failures := make([]error, 10)
	var group sync.WaitGroup
	for index := range failures {
		group.Go(func() { failures[index] = ledger.Transfer("A", "B", 1) })
	}
	group.Wait()
	snapshot := ledger.Snapshot()
	fmt.Printf("A=%d B=%d 총합=%d\n", snapshot["A"], snapshot["B"], snapshot["A"]+snapshot["B"])
	fmt.Println("모두 성공:", errors.Join(failures...) == nil)
}
