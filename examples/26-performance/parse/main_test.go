package main

import (
	"errors"
	"strconv"
	"testing"
)

func TestSumNumbers(t *testing.T) {
	for _, tc := range []struct {
		values []string
		want   int
		bad    bool
	}{
		{nil, 0, false}, {[]string{"10", "-3"}, 7, false},
		{[]string{"10", "x"}, 0, true}, {[]string{""}, 0, true},
	} {
		got, err := sumNumbers(tc.values)
		if (err != nil) != tc.bad || got != tc.want {
			t.Fatalf("got %d, %v", got, err)
		}
		if tc.bad && !errors.Is(err, strconv.ErrSyntax) {
			t.Fatalf("lost cause: %v", err)
		}
	}
}

func BenchmarkSumNumbers(b *testing.B) {
	values := []string{"10", "20", "30", "40"}
	b.SetBytes(int64(len("10203040")))
	b.ReportAllocs()
	for b.Loop() {
		if _, err := sumNumbers(values); err != nil {
			b.Fatal(err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 합계: 60
}
