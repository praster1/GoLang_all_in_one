package main

import "fmt"

type Counter struct{ value int }

func (counter Counter) Value() int { return counter.value }
func (counter *Counter) Add(amount int) bool {
	if counter == nil || counter.value < 0 || counter.value > 100 || amount < 0 || amount > 100-counter.value {
		return false
	}
	counter.value += amount
	return true
}
func (counter *Counter) Reset() bool {
	if counter == nil {
		return false
	}
	counter.value = 0
	return true
}

func main() {
	counter := Counter{}
	ok := counter.Add(60)
	fmt.Println("추가:", ok, counter.Value())
	ok = counter.Add(41)
	fmt.Println("한도 초과:", ok, counter.Value())
	ok = counter.Reset()
	fmt.Println("초기화:", ok, counter.Value())
}
