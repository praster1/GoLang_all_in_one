package main

import "testing"

func TestNewStepperSequence(t *testing.T) {
	tests := []struct {
		name  string
		start int
		step  int
		want  []int
	}{
		{"positive step", 0, 1, []int{1, 2}},
		{"zero step", 10, 0, []int{10, 10}},
		{"negative step", 5, -2, []int{3, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := newStepper(tt.start, tt.step)
			for call, want := range tt.want {
				if got := next(); got != want {
					t.Errorf("newStepper(%d, %d), call %d = %d, want %d", tt.start, tt.step, call+1, got, want)
				}
			}
		})
	}
}

func TestNewStepperIndependentState(t *testing.T) {
	first := newStepper(10, 3)
	second := newStepper(100, 10)
	if got := first(); got != 13 {
		t.Errorf("first call = %d, want 13", got)
	}
	if got := first(); got != 16 {
		t.Errorf("second call = %d, want 16", got)
	}
	if got := second(); got != 110 {
		t.Errorf("other stepper first call = %d, want 110", got)
	}
	if got := first(); got != 19 {
		t.Errorf("first stepper after other call = %d, want 19", got)
	}
	if got := second(); got != 120 {
		t.Errorf("other stepper second call = %d, want 120", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 첫 번째: 13
	// 첫 번째: 16
	// 두 번째: 110
	// 첫 번째: 19
}
