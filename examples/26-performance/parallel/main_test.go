package main

import (
	"sync/atomic"
	"testing"
)

func TestCount(t *testing.T) {
	if count(0, 10) != 0 || count(2, 0) != 0 || count(4, 100) != 400 {
		t.Fatal("wrong total")
	}
}

func BenchmarkAtomicParallel(b *testing.B) {
	var total atomic.Int64
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			total.Add(1)
		}
	})
	if total.Load() != int64(b.N) {
		b.Fatal("lost update")
	}
}

func Example_main() {
	main()
	// Output:
	// 처리: 400
}
