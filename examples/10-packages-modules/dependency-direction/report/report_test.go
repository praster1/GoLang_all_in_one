package report

import "testing"

type fixed int

func (value fixed) Total() int { return int(value) }
func TestSummary(t *testing.T) {
	if got := Summary(fixed(42)); got != "합계: 42원" {
		t.Fatal(got)
	}
}
