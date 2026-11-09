package main

import (
	"errors"
	"strings"
	"testing"
)

func TestSumCSV(t *testing.T) {
	cases := []struct {
		input string
		want  int
	}{
		{"name,minutes\n", 0},
		{"name,minutes\n休,0\n", 0},
		{"name,minutes\r\n\"Go, basics\",25\r\n테스트,35", 60},
		{"name,minutes\n" + strings.Repeat("Go,120\n", 1000), 120000},
	}
	for _, test := range cases {
		got, err := sumCSV(strings.NewReader(test.input))
		if err != nil || got != test.want {
			t.Fatalf("sumCSV = %d, %v; want %d", got, err, test.want)
		}
	}
	for _, input := range []string{
		"", "name,minute\n", "name,minutes\nGo\n", "name,minutes\nGo,1,2\n",
		"name,minutes\nGo,25\nbad,x\n", "name,minutes\nGo,-1\n",
		"name,minutes\nGo,121\n", "name,minutes\n ,10\n",
		"name,minutes\n" + strings.Repeat("Go,1\n", 1001),
	} {
		got, err := sumCSV(strings.NewReader(input))
		if err == nil || got != 0 {
			t.Errorf("invalid input => %d, %v", got, err)
		}
	}
}

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestSumCSVReadFailure(t *testing.T) {
	want := errors.New("read failed")
	_, err := sumCSV(failedReader{want})
	if !errors.Is(err, want) {
		t.Fatal(err)
	}
}

func Example_main() {
	main()
	// Output:
	// 합계: 60분
}
