package main

import "fmt"

type Reader interface{ Read() int }
type Meter struct{ Value int }

func (meter Meter) Read() int       { return meter.Value }
func (meter *Meter) Add(amount int) { meter.Value += amount }

type NamedMeter struct {
	Meter
	Name string
}

func main() {
	meter := NamedMeter{Meter: Meter{Value: 2}, Name: "학습"}
	meter.Add(3)
	var reader Reader = meter
	fmt.Printf("%s: %d\n", meter.Name, reader.Read())
	fmt.Println("명시적 내부 접근:", meter.Meter.Read())
}
