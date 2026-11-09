package main

import "testing"

func TestChecksum(t *testing.T) {
	if got := checksum([]byte("Go"), 2); got != 2674860 {
		t.Fatalf("got %d", got)
	}
	if checksum(nil, 100) != 0 || checksum([]byte("Go"), 0) != 0 {
		t.Fatal("empty work")
	}
}

func BenchmarkChecksum(b *testing.B) {
	data := make([]byte, 4096)
	for i := range data {
		data[i] = byte(i)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(data) * 50))
	for b.Loop() {
		checksum(data, 50)
	}
}

func Example_main() {
	main()
	// Output:
	// 체크섬: 2674860
}
