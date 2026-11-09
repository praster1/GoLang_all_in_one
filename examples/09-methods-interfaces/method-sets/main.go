package main

import "fmt"

type Reader interface{ Read() int }
type Adder interface{ Add(int) }

type Counter struct{ N int }

func (counter Counter) Read() int       { return counter.N }
func (counter *Counter) Add(amount int) { counter.N += amount }

var _ Reader = Counter{}
var _ Reader = (*Counter)(nil)
var _ Adder = (*Counter)(nil)

func main() {
	counter := Counter{N: 2}
	var snapshot Reader = counter
	var live Reader = &counter
	var adder Adder = &counter
	adder.Add(3)
	fmt.Println("값을 담은 인터페이스:", snapshot.Read())
	fmt.Println("포인터를 담은 인터페이스:", live.Read())
	counter.Add(1)
	fmt.Println("직접 호출 뒤:", counter.Read())
}
