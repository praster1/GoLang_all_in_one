package main

import "testing"

func TestGrade(t *testing.T) {
	got, err := grade(80)
	if err != nil {
		t.Fatalf("grade(80) error: %v", err)
	}
	if got != "우수" {
		t.Errorf("grade(80) = %q; want %q", got, "우수")
	}
}

func TestGradeInvalid(t *testing.T) {
	got, err := grade(-1)
	if err == nil {
		t.Fatal("grade(-1) accepted an invalid score")
	}
	if got != "" {
		t.Errorf("result on error = %q; want empty", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 판정: 우수
}
