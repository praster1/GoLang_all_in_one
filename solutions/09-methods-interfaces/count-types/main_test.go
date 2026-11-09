package main

import "testing"

type customInt int

func TestCountTypes(t *testing.T) {
	for _, tc := range []struct {
		values []any
		want   Counts
	}{
		{nil, Counts{}}, {[]any{0, "", false}, Counts{Ints: 1, Strings: 1, Bools: 1}},
		{[]any{int64(1), customInt(1), (*int)(nil), []int{1}, nil}, Counts{Others: 5}},
	} {
		if got := countTypes(tc.values); got != tc.want {
			t.Fatalf("%v: %+v", tc.values, got)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// int: 1, string: 2, bool: 2, 기타: 2
}
