package main

import "testing"

func TestHits(t *testing.T) {
	for _, tc := range []struct {
		values, queries []int
		want            int
	}{
		{nil, []int{0}, 0}, {[]int{1}, nil, 0},
		{[]int{1, 1, 2}, []int{1, 1, 3}, 2}, {[]int{2, 4, 6}, []int{4, 5, 4}, 2},
	} {
		if hitsLinear(tc.values, tc.queries) != tc.want || hitsIndexed(index(tc.values), tc.queries) != tc.want {
			t.Fatal("incorrect hits")
		}
	}
}

func BenchmarkQueryBatch(b *testing.B) {
	values := make([]int, 1000)
	for i := range values {
		values[i] = i
	}
	queries := []int{0, 500, 999, 1001}
	set := index(values)
	b.Run("Linear", func(b *testing.B) {
		for b.Loop() {
			hitsLinear(values, queries)
		}
	})
	b.Run("Prepared", func(b *testing.B) {
		for b.Loop() {
			hitsIndexed(set, queries)
		}
	})
	b.Run("BuildIncluded", func(b *testing.B) {
		for b.Loop() {
			hitsIndexed(index(values), queries)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 선형: 2
	// 인덱스: 2
}
