package main

import (
	"bytes"
	"encoding/csv"
	"errors"
	"reflect"
	"testing"
)

type failedWriter struct{ err error }

func (writer failedWriter) Write([]byte) (int, error) { return 0, writer.err }

func TestWriteSessionsRoundTrip(t *testing.T) {
	sessions := []Session{{"Go, basics", 25}, {"한\n줄", 35}, {`따옴표 "하나"`, 1}}
	var output bytes.Buffer
	if err := writeSessions(&output, sessions); err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(&output).ReadAll()
	want := [][]string{{"name", "minutes"}, {"Go, basics", "25"}, {"한\n줄", "35"}, {`따옴표 "하나"`, "1"}}
	if err != nil || !reflect.DeepEqual(records, want) {
		t.Fatalf("records = %#v, %v", records, err)
	}
}

func TestWriteSessionsFlushError(t *testing.T) {
	want := errors.New("write failure")
	if err := writeSessions(failedWriter{want}, nil); !errors.Is(err, want) {
		t.Fatalf("error = %v", err)
	}
}

func Example_main() {
	main()
	// Output:
	// name,minutes
	// "Go, basics",25
	// "문자열 ""연습""",35
}
