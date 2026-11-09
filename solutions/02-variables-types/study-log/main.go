package main

import "fmt"

func main() {
	var course string = "Go"
	days := 2
	days += 1
	minutesPerDay := 25
	totalMinutes := days * minutesPerDay
	ready := true

	fmt.Printf("%s 학습\n", course)
	fmt.Printf("%d일, 하루 %d분\n", days, minutesPerDay)
	fmt.Printf("총 %d분\n", totalMinutes)
	fmt.Printf("준비: %t\n", ready)
}
