package main

import "testing"

func TestTotalMinutes(t *testing.T) {
	tests := []struct {
		name          string
		days          int
		minutesPerDay int
		wantTotal     int
		wantOK        bool
	}{
		{"zero days", 0, 25, 0, true},
		{"smallest positive plan", 1, 1, 1, true},
		{"default input", 5, 25, 125, true},
		{"maximum plan", 7, 120, 840, true},
		{"negative days", -1, 25, 0, false},
		{"too many days", 8, 25, 0, false},
		{"zero daily minutes", 5, 0, 0, false},
		{"too many daily minutes", 5, 121, 0, false},
		{"zero days still checks daily minutes", 0, 0, 0, false},
		{"zero days rejects excess daily minutes", 0, 121, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total, ok := totalMinutes(tt.days, tt.minutesPerDay)
			if total != tt.wantTotal || ok != tt.wantOK {
				t.Errorf("totalMinutes(%d, %d) = (%d, %t), want (%d, %t)", tt.days, tt.minutesPerDay, total, ok, tt.wantTotal, tt.wantOK)
			}
		})
	}
}

func TestSplitMinutes(t *testing.T) {
	tests := []struct {
		total       int
		wantHours   int
		wantMinutes int
	}{
		{0, 0, 0},
		{59, 0, 59},
		{60, 1, 0},
		{125, 2, 5},
		{840, 14, 0},
	}
	for _, tt := range tests {
		hours, minutes := splitMinutes(tt.total)
		if hours != tt.wantHours || minutes != tt.wantMinutes {
			t.Errorf("splitMinutes(%d) = (%d, %d), want (%d, %d)", tt.total, hours, minutes, tt.wantHours, tt.wantMinutes)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 총 계획: 125분
	// 환산: 2시간 5분
}
