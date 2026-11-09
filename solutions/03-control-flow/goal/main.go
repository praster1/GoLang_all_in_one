package main

import "fmt"

func main() {
	targetMinutes := 100
	dailyMinutes := 30

	if targetMinutes < 0 || dailyMinutes <= 0 {
		fmt.Println("입력 확인")
	} else {
		days := 0
		total := 0
		for total < targetMinutes {
			total += dailyMinutes
			days++
		}
		fmt.Printf("%d일, 누적 %d분\n", days, total)
	}
}
