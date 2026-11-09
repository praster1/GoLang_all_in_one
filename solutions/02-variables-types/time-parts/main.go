package main

import "fmt"

func main() {
	const secondsPerMinute = 60
	const minutesPerHour = 60
	totalSeconds := 9865

	hours := totalSeconds / (secondsPerMinute * minutesPerHour)
	minutes := totalSeconds / secondsPerMinute % minutesPerHour
	seconds := totalSeconds % secondsPerMinute

	fmt.Printf("%d초 = %d시간 %d분 %d초\n", totalSeconds, hours, minutes, seconds)
}
