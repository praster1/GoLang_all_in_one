package main

import (
	"strings"
	"testing"
)

func TestRecord(t *testing.T) {
	for _, input := range []string{"", "=x", "A=x", "a-b=x", "a=x\ny"} {
		if _, _, err := parseRecord(input); err == nil {
			t.Fatalf("accepted %q", input)
		}
	}
	key, value, err := parseRecord("a_b=x=y")
	if err != nil || key != "a_b" || value != "x=y" {
		t.Fatal("split contract")
	}
}

func FuzzRecord(f *testing.F) {
	for _, seed := range []string{"", "a=", "title=Go=입문", "=x", "a=x\ny"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, input string) {
		if len(input) > 8192 {
			t.Skip()
		}
		key, value, err := parseRecord(input)
		if err != nil {
			return
		}
		if key+"="+value != input {
			t.Fatal("lost bytes")
		}
		k, v, e := parseRecord(key + "=" + value)
		if e != nil || k != key || v != value {
			t.Fatal("round trip")
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 키: title
	// 값: Go=입문
}

func TestRecordLengthBoundary(t *testing.T) {
	for _, n := range []int{4094, 4095} {
		_, _, err := parseRecord("a=" + strings.Repeat("x", n))
		if (err != nil) != (n == 4095) {
			t.Fatalf("value bytes=%d err=%v", n, err)
		}
	}
}
