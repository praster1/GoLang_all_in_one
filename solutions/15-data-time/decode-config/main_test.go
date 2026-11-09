package main

import (
	"errors"
	"strings"
	"testing"
)

func TestReadConfig(t *testing.T) {
	valid := `{"course":" Go ","daily_minutes":25}`
	for _, input := range []string{valid, valid + strings.Repeat(" ", 4096-len(valid))} {
		config, err := readConfig(strings.NewReader(input))
		if err != nil || config != (Config{"Go", 25}) {
			t.Fatalf("readConfig = %#v, %v", config, err)
		}
	}
	for _, input := range []string{
		"", "{", "null", "{}", `{"course":" ","daily_minutes":25}`,
		`{"course":"Go","daily_minutes":0}`, `{"course":"Go","daily_minutes":121}`,
		`{"course":"Go","daily_minutes":1.5}`,
		`{"course":"Go","daily_minutes":25,"extra":1}`,
		valid + "{}", valid + "x", valid + strings.Repeat(" ", 4097-len(valid)),
	} {
		config, err := readConfig(strings.NewReader(input))
		if err == nil || config != (Config{}) {
			t.Errorf("invalid %q => %#v, %v", input, config, err)
		}
	}
	for _, minutes := range []string{"1", "120"} {
		if _, err := readConfig(strings.NewReader(`{"course":"Go","daily_minutes":` + minutes + `}`)); err != nil {
			t.Fatal(err)
		}
	}
}

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestReadConfigFailure(t *testing.T) {
	want := errors.New("read failed")
	_, err := readConfig(failedReader{want})
	if !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func Example_main() {
	main()
	// Output:
	// Go: 하루 25분
}
