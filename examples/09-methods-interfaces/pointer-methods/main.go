package main

import "fmt"

type Counter struct{ Value int }

func (counter *Counter) Add(amount int) bool {
	if counter == nil || counter.Value < 0 || counter.Value > 100 || amount < 0 || amount > 100-counter.Value {
		return false
	}
	counter.Value += amount
	return true
}

func main() {
	counter := Counter{}
	ok := counter.Add(3)
	fmt.Println("값 변수 호출:", ok, counter.Value)
	pointer := &counter
	ok = pointer.Add(2)
	fmt.Println("포인터 호출:", ok, counter.Value)
	var missing *Counter
	fmt.Println("nil 수신자:", missing.Add(1))
}
