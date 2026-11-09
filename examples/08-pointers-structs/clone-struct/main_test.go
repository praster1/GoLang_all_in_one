package main

import (
	"slices"
	"testing"
)

func TestCloneProfile(t *testing.T) {
	original := Profile{Name: "Mina", Scores: []int{70, 80}, Labels: map[string]string{"team": "go"}}
	cloned := cloneProfile(original)
	if cloned.Name != original.Name || !slices.Equal(cloned.Scores, original.Scores) || cloned.Labels["team"] != "go" {
		t.Fatalf("copy lost data: %+v", cloned)
	}
	cloned.Name = "Jin"
	cloned.Scores[0] = 99
	cloned.Labels["team"] = "web"
	if original.Name != "Mina" || original.Scores[0] != 70 || original.Labels["team"] != "go" {
		t.Fatalf("shared data: %+v", original)
	}
	zero := cloneProfile(Profile{})
	if zero.Scores != nil || zero.Labels != nil {
		t.Fatalf("nil changed: %+v", zero)
	}
	empty := cloneProfile(Profile{Scores: []int{}, Labels: map[string]string{}})
	if empty.Scores == nil || empty.Labels == nil {
		t.Fatal("empty changed to nil")
	}
	empty.Labels["new"] = "value"
}

func Example_main() {
	main()
	// Output:
	// 얕은 복사 뒤 원본: Mina, [99 80], web
	// 독립 복사 뒤 원본: Mina, [99 80], web
	// 독립 복사본: Sol, [7 80], cli
}
