package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestConvertCSV(t *testing.T) {
	var output bytes.Buffer
	input := "name,minutes\n\"Go, basics\",25\n休,0\n"
	if err := convertCSV(strings.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	var sessions []Session
	if err := json.Unmarshal(output.Bytes(), &sessions); err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 2 || sessions[0] != (Session{"Go, basics", 25}) || sessions[1].Minutes != 0 {
		t.Fatalf("sessions = %#v", sessions)
	}
	output.Reset()
	if err := convertCSV(strings.NewReader("name,minutes\n"), &output); err != nil || output.String() != "[]\n" {
		t.Fatalf("header only = %q, %v", output.String(), err)
	}
}

func TestInvalidCSVDoesNotWrite(t *testing.T) {
	for _, input := range []string{
		"", "name,minute\n", "name,minutes\nGo\n", "name,minutes\nGo,1,2\n",
		"name,minutes\nGo,25\nbad,x\n", "name,minutes\nGo,-1\n",
		"name,minutes\nGo,121\n", "name,minutes\n ,25\n",
		"name,minutes\n" + strings.Repeat("Go,1\n", 1001),
	} {
		var output bytes.Buffer
		err := convertCSV(strings.NewReader(input), &output)
		if err == nil || output.Len() != 0 {
			t.Errorf("invalid input wrote %q, error %v", output.String(), err)
		}
	}
	var output bytes.Buffer
	if err := convertCSV(strings.NewReader("name,minutes\n"+strings.Repeat("Go,120\n", 1000)), &output); err != nil {
		t.Fatal(err)
	}
}

type failedWriter struct{ err error }

func (writer failedWriter) Write([]byte) (int, error) { return 0, writer.err }

type failedReader struct{ err error }

func (reader failedReader) Read([]byte) (int, error) { return 0, reader.err }

func TestConversionIOFailures(t *testing.T) {
	want := errors.New("I/O failure")
	if err := convertCSV(strings.NewReader("name,minutes\n"), failedWriter{want}); !errors.Is(err, want) {
		t.Fatalf("writer error = %v", err)
	}
	var output bytes.Buffer
	if err := convertCSV(failedReader{want}, &output); !errors.Is(err, want) || output.Len() != 0 {
		t.Fatalf("reader error = %v, output = %q", err, output.String())
	}
}

func Example_main() {
	main()
	// Output:
	// [
	//   {
	//     "name": "Go",
	//     "minutes": 25
	//   },
	//   {
	//     "name": "테스트",
	//     "minutes": 35
	//   }
	// ]
}
