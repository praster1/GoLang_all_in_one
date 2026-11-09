package main

import "testing"

func TestFirstBytes(t *testing.T) {
	for _, tc := range []struct {
		n    int
		want string
	}{{-1, ""}, {0, ""}, {2, "Go"}, {100, "Go!"}} {
		input := []byte("Go!")
		got := firstBytes(input, tc.n)
		if string(got) != tc.want {
			t.Fatalf("n=%d got=%q", tc.n, got)
		}
		if len(got) > 0 {
			got[0] = 'X'
			if input[0] != 'G' {
				t.Fatal("shared storage")
			}
		}
	}
}

func BenchmarkFirstBytes(b *testing.B) {
	data := make([]byte, 1024*1024)
	b.ReportAllocs()
	for b.Loop() {
		firstBytes(data, 32)
	}
}

func Example_main() {
	main()
	// Output:
	// Go
	// No language
}
