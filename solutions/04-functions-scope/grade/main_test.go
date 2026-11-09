package main

import "testing"

func TestClassifyScore(t *testing.T) {
	tests := []struct {
		name      string
		score     int
		wantLabel string
		wantOK    bool
	}{
		{"below minimum", -1, "", false},
		{"minimum", 0, "다시 연습", true},
		{"below passing", 59, "다시 연습", true},
		{"passing boundary", 60, "통과", true},
		{"default input", 72, "통과", true},
		{"below excellent", 89, "통과", true},
		{"excellent boundary", 90, "우수", true},
		{"maximum", 100, "우수", true},
		{"above maximum", 101, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label, ok := classifyScore(tt.score)
			if label != tt.wantLabel || ok != tt.wantOK {
				t.Errorf("classifyScore(%d) = (%q, %t), want (%q, %t)", tt.score, label, ok, tt.wantLabel, tt.wantOK)
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// 72점: 통과
}
