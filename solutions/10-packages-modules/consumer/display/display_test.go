package display

import "testing"

type text string

func (v text) Description() string { return string(v) }
func TestLine(t *testing.T) {
	if got := Line(text("새 구현")); got != "- 새 구현" {
		t.Fatal(got)
	}
}
