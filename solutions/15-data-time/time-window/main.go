package main

import (
	"fmt"
	"time"
)

func withinWindow(value, start, end time.Time) bool {
	return start.Before(end) && !value.Before(start) && value.Before(end)
}

func main() {
	start := time.Date(2026, time.November, 9, 9, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	values := []time.Time{
		start.Add(-time.Nanosecond),
		start,
		end.Add(-time.Nanosecond),
		end,
	}
	for _, value := range values {
		fmt.Println(value.Format(time.RFC3339Nano), withinWindow(value, start, end))
	}
}
