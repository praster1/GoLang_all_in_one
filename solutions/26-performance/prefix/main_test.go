package main

import (
	"errors"
	"testing"
)

func TestClonePrefix(t *testing.T) {
	for _, tc := range []struct {
		n    int
		want string
		bad  bool
	}{{-1, "", true}, {0, "", false}, {2, "Go", false}, {3, "Go!", false}, {4, "", true}} {
		input := []byte("Go!")
		got, err := clonePrefix(input, tc.n)
		if errors.Is(err, errLength) != tc.bad || string(got) != tc.want {
			t.Fatalf("n=%d got=%q err=%v", tc.n, got, err)
		}
		if len(got) > 0 {
			got[0] = 'X'
			if input[0] != 'G' {
				t.Fatal("alias")
			}
		}
	}
}

func BenchmarkClonePrefix(b *testing.B) {
	data := make([]byte, 4096)
	for _, tc := range []struct {
		name string
		n    int
	}{{"N32", 32}, {"N256", 256}, {"N4096", 4096}} {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(tc.n))
			for b.Loop() {
				if _, err := clonePrefix(data, tc.n); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func Example_main() {
	main()
	// Output:
	// prefix=Go
	// source=No!
}
