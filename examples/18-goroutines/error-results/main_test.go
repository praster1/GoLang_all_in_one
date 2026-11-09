package main

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestParseAll(t *testing.T) {
	for _, c := range []struct {
		input []string
		want  []int
	}{{nil, []int{}}, {[]string{"3", "1", "2"}, []int{3, 1, 2}}, {[]string{"0", "-1"}, []int{0, -1}}} {
		got, err := parseAll(c.input)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatal(got, err)
		}
	}
	got, err := parseAll([]string{"x", "2", "y"})
	var numberError *strconv.NumError
	if got != nil || !errors.As(err, &numberError) || !strings.Contains(err.Error(), "항목 0") || !strings.Contains(err.Error(), "항목 2") {
		t.Fatal(got, err)
	}
	if got, err := parseAll(make([]string, 17)); got != nil || err == nil {
		t.Fatal(got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 변환 실패: true
	// 실패 결과 없음: true
	// 숫자 오류 확인: true
}
