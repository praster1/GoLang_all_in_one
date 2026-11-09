package main

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func FuzzHexRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 15, 255})
	f.Add([]byte("한글"))
	f.Fuzz(func(t *testing.T, input []byte) {
		if len(input) > 65536 {
			t.Skip()
		}
		encoded := encode(input)
		if len(encoded) != 2*len(input) {
			t.Fatal("wrong encoded length")
		}
		decoded, err := hex.DecodeString(encoded)
		if err != nil || !bytes.Equal(decoded, input) {
			t.Fatalf("round trip: %v", err)
		}
	})
}

func TestHexInvalid(t *testing.T) {
	for _, input := range []string{"0", "zz", "0x12"} {
		if _, err := hex.DecodeString(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 000fff
	// [0 15 255]
}
