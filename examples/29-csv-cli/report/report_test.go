package report

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseRow(t *testing.T) {
	for _, tc := range []struct {
		name   string
		fields []string
		want   Record
		bad    bool
	}{
		{"trim", []string{" 도서 ", " 0012 "}, Record{"도서", 12}, false},
		{"zero", []string{"도서", "0"}, Record{"도서", 0}, false},
		{"max", []string{"도서", "1000000000"}, Record{"도서", 1000000000}, false},
		{"shape", []string{"도서"}, Record{}, true},
		{"empty", []string{" ", "1"}, Record{}, true},
		{"max category", []string{strings.Repeat("가", 40), "1"}, Record{strings.Repeat("가", 40), 1}, false},
		{"long", []string{strings.Repeat("가", 41), "1"}, Record{}, true},
		{"utf8", []string{string([]byte{0xff}), "1"}, Record{}, true},
		{"control", []string{"a\nb", "1"}, Record{}, true},
		{"negative", []string{"x", "-1"}, Record{}, true},
		{"plus", []string{"x", "+1"}, Record{}, true},
		{"decimal", []string{"x", "1.0"}, Record{}, true},
		{"over", []string{"x", "1000000001"}, Record{}, true},
		{"overflow", []string{"x", "99999999999999999999999"}, Record{}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseRow(tc.fields, 7)
			if (err != nil) != tc.bad || got != tc.want {
				t.Fatalf("got=%v err=%v", got, err)
			}
			if tc.bad {
				var detail *DataError
				if !errors.Is(err, ErrInput) || !errors.As(err, &detail) || detail.Record != 7 {
					t.Fatalf("error contract: %v", err)
				}
			}
		})
	}
}

func TestAnalyze(t *testing.T) {
	text := "category,amount\n도서,12000\n식품,5000\n도서,8000\n"
	got, err := Analyze(strings.NewReader(text), DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	want := []Group{{"도서", 2, 20000}, {"식품", 1, 5000}}
	if got.Rows != 3 || got.Total != 25000 || !slices.Equal(got.Groups, want) {
		t.Fatalf("got=%+v", got)
	}
	empty, err := Analyze(strings.NewReader("category,amount\n"), DefaultOptions())
	if err != nil || empty.Rows != 0 || empty.Total != 0 || empty.Groups == nil {
		t.Fatal("empty summary")
	}
	quoted, err := Analyze(strings.NewReader("category,amount\r\n\"개발,도구\",3\r\n"), DefaultOptions())
	if err != nil || quoted.Groups[0].Category != "개발,도구" {
		t.Fatal("CSV quoting")
	}
}

func TestAnalyzeLimitsAndFailures(t *testing.T) {
	text := "category,amount\nx,1\n"
	opt := Options{MaxBytes: int64(len(text)), MaxRows: 1}
	if _, err := Analyze(strings.NewReader(text), opt); err != nil {
		t.Fatal(err)
	}
	opt.MaxBytes--
	if _, err := Analyze(strings.NewReader(text), opt); !errors.Is(err, ErrLimit) {
		t.Fatalf("byte limit: %v", err)
	}
	opt = Options{MaxBytes: 1000, MaxRows: 1}
	if _, err := Analyze(strings.NewReader(text+"x,2\n"), opt); !errors.Is(err, ErrLimit) {
		t.Fatalf("row limit: %v", err)
	}
	for _, input := range []string{"", "wrong,amount\nx,1\n", "category,amount\nx,1,extra\n", "category,amount\nx,bad\n", "category,amount\n\"broken,1\n"} {
		got, err := Analyze(strings.NewReader(input), DefaultOptions())
		if err == nil || got.Rows != 0 || got.Total != 0 || len(got.Groups) != 0 {
			t.Fatalf("partial result: %+v %v", got, err)
		}
	}
	for _, opt := range []Options{{0, 1}, {1, 0}, {1 << 30, 1}, {1, 1000001}} {
		if _, err := Analyze(strings.NewReader(text), opt); err == nil {
			t.Fatal("invalid options accepted")
		}
	}
}

func TestAnalyzeByteLimitAtEOF(t *testing.T) {
	// CSV skips blank lines, but those bytes still count toward the input limit.
	for _, text := range []string{
		"category,amount\n",
		"category,amount\n\n\n",
		"category,amount\n한글,1\n\n\n",
	} {
		t.Run(fmt.Sprintf("bytes=%d", len(text)), func(t *testing.T) {
			exact := Options{MaxBytes: int64(len(text)), MaxRows: 1}
			got, err := Analyze(strings.NewReader(text), exact)
			if err != nil || got.Groups == nil {
				t.Fatalf("exact limit: summary=%+v err=%v", got, err)
			}
			exact.MaxBytes--
			got, err = Analyze(strings.NewReader(text), exact)
			if !errors.Is(err, ErrLimit) || got.Rows != 0 || got.Total != 0 || len(got.Groups) != 0 {
				t.Fatalf("one byte over: summary=%+v err=%v", got, err)
			}
		})
	}
}

type failedReader struct{ err error }

func (r failedReader) Read(p []byte) (int, error) { return 0, r.err }

type failedWriter struct{ err error }

func (w failedWriter) Write(p []byte) (int, error) { return 0, w.err }

func TestIOErrors(t *testing.T) {
	sentinel := errors.New("device failure")
	if _, err := Analyze(failedReader{sentinel}, DefaultOptions()); !errors.Is(err, sentinel) {
		t.Fatal("read cause lost")
	}
	partial := io.MultiReader(strings.NewReader("category,amount\nx,1\n"), failedReader{sentinel})
	got, err := Analyze(partial, DefaultOptions())
	if !errors.Is(err, sentinel) || got.Rows != 0 || got.Total != 0 || len(got.Groups) != 0 {
		t.Fatalf("partial read result: %+v, %v", got, err)
	}
	for _, write := range []func(io.Writer, Summary) error{WriteCSV, WriteJSON} {
		if err := write(failedWriter{sentinel}, Summary{}); !errors.Is(err, sentinel) {
			t.Fatal("write cause lost")
		}
	}
}

func TestFormatsAndInputPreservation(t *testing.T) {
	summary := Summary{Rows: 2, Total: 3, Groups: []Group{{"z", 1, 1}, {"a,b", 1, 2}}}
	before := slices.Clone(summary.Groups)
	var csvOut, jsonOut bytes.Buffer
	if err := WriteCSV(&csvOut, summary); err != nil {
		t.Fatal(err)
	}
	if csvOut.String() != "category,count,total\n\"a,b\",1,2\nz,1,1\n" {
		t.Fatal(csvOut.String())
	}
	if err := WriteJSON(&jsonOut, summary); err != nil {
		t.Fatal(err)
	}
	var decoded Summary
	if err := json.Unmarshal(jsonOut.Bytes(), &decoded); err != nil || decoded.Rows != 2 || decoded.Total != 3 || decoded.Groups[0].Category != "a,b" {
		t.Fatal("JSON contract")
	}
	if !slices.Equal(before, summary.Groups) {
		t.Fatal("writer mutated source")
	}
}

func FuzzAnalyze(f *testing.F) {
	for _, seed := range []string{"", "category,amount\n", "category,amount\n한글,0\n", "category,amount\n\"a,b\",10\n"} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 8192 {
			t.Skip()
		}
		got, err := Analyze(bytes.NewReader(data), Options{MaxBytes: 4096, MaxRows: 100})
		if err != nil {
			return
		}
		rows := 0
		var total int64
		for i, group := range got.Groups {
			if group.Count < 1 || group.Total < 0 || !utf8.ValidString(group.Category) {
				t.Fatal("invalid group")
			}
			if i > 0 && got.Groups[i-1].Category >= group.Category {
				t.Fatal("order or duplicate")
			}
			rows += group.Count
			total += group.Total
		}
		if rows != got.Rows || total != got.Total || rows > 100 || total > 100000000000 {
			t.Fatal("summary invariant")
		}
	})
}
