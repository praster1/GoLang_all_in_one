package main

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
	"testing"
)

func FuzzCSVPair(f *testing.F) {
	f.Add("", "")
	f.Add("Go,입문", "한글")
	f.Add("a\nb", "quote \"here\"")
	f.Add("a\r\nb", "x")
	f.Fuzz(func(t *testing.T, left, right string) {
		if len(left)+len(right) > 65536 {
			t.Skip()
		}
		encoded, err := encodePair(left, right)
		if strings.ContainsAny(left+right, "\r") {
			if !errors.Is(err, errCarriageReturn) {
				t.Fatal("CR accepted")
			}
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		r := csv.NewReader(strings.NewReader(encoded))
		row, err := r.Read()
		if err != nil || len(row) != 2 || row[0] != left || row[1] != right {
			t.Fatalf("round trip: %q %v", row, err)
		}
		if _, err = r.Read(); err != io.EOF {
			t.Fatalf("extra record: %v", err)
		}
	})
}

func Example_main() {
	main()
	// Output:
	// "Go,입문",한글
}
