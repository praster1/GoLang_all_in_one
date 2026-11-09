package main

import "fmt"

func totalMinutes(days, minutesPerDay int) (int, bool) {
	if days < 0 || days > 7 || minutesPerDay < 1 || minutesPerDay > 120 {
		return 0, false
	}
	return days * minutesPerDay, true
}

func splitMinutes(total int) (int, int) {
	return total / 60, total % 60
}

func main() {
	days := 5
	minutesPerDay := 25
	total, ok := totalMinutes(days, minutesPerDay)
	if !ok {
		fmt.Println("계획 확인")
		return
	}
	hours, minutes := splitMinutes(total)
	fmt.Printf("총 계획: %d분\n환산: %d시간 %d분\n", total, hours, minutes)
}
