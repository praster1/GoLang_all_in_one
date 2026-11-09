package main

// mean accepts at most 100 scores, each in 0..100.
func mean(scores []int) (float64, bool) {
	if len(scores) == 0 || len(scores) > 100 {
		return 0, false
	}
	total := 0
	for _, score := range scores {
		if score < 0 || score > 100 {
			return 0, false
		}
		total += score
	}
	return float64(total) / float64(len(scores)), true
}
