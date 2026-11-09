package main

// totalMinutes adds a small list of nonnegative lesson durations.
func totalMinutes(values []int) int {
	total := 0
	for _, value := range values {
		total += value
	}
	return total
}
