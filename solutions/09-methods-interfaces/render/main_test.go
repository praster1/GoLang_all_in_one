package main

import (
	"slices"
	"testing"
)

type testDescriber struct {
	calls *int
	text  string
}

func (value testDescriber) Describe() string { *value.calls++; return value.text }
func TestRender(t *testing.T) {
	calls := 0
	custom := testDescriber{calls: &calls, text: "custom"}
	got, ok := render([]Describer{custom, Number(2), (*Note)(nil)})
	if !ok || !slices.Equal(got, []string{"custom", "숫자: 2", "메모 없음"}) || calls != 1 {
		t.Fatalf("%q %t calls=%d", got, ok, calls)
	}
	calls = 0
	got, ok = render([]Describer{custom, nil})
	if ok || got != nil || calls != 0 {
		t.Fatalf("failure should call nothing: %q %t %d", got, ok, calls)
	}
	got, ok = render(nil)
	if !ok || got != nil {
		t.Fatal("nil slice contract")
	}
}

func Example_main() {
	main()
	// Output:
	// 성공: true
	// 메모: Go
	// 숫자: 7
	// 메모 없음
	// nil 인터페이스: false
}
