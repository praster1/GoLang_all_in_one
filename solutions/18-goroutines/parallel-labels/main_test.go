package main

import (
	"reflect"
	"testing"
)

func TestLabels(t *testing.T) {
	for _, c := range []struct{ input, want []string }{{nil, []string{}}, {[]string{" go ", "test", "한글"}, []string{"GO", "TEST", "한글"}}, {[]string{"same", "same"}, []string{"SAME", "SAME"}}} {
		got, err := labels(c.input)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatal(got, err)
		}
	}
	maximum := make([]string, 16)
	for i := range maximum {
		maximum[i] = "go"
	}
	if got, err := labels(maximum); err != nil || len(got) != 16 {
		t.Fatal(got, err)
	}
	for _, input := range [][]string{{""}, {"go", " "}, make([]string, 17)} {
		got, err := labels(input)
		if got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// [GO TEST 한글]
}
