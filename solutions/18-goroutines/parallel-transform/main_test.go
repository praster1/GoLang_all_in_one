package main

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestTransformAll(t *testing.T) {
	upper := func(ctx context.Context, input string) (string, error) { return strings.ToUpper(input), ctx.Err() }
	for _, c := range []struct{ input, want []string }{{nil, []string{}}, {[]string{"b", "a"}, []string{"B", "A"}}} {
		got, err := transformAll(context.Background(), c.input, upper)
		if err != nil || !reflect.DeepEqual(got, c.want) {
			t.Fatal(got, err)
		}
	}
	maximum := make([]string, 16)
	if got, err := transformAll(context.Background(), maximum, upper); err != nil || len(got) != 16 {
		t.Fatal(got, err)
	}
	for _, c := range []struct {
		input []string
		fn    func(context.Context, string) (string, error)
	}{{nil, nil}, {make([]string, 17), upper}} {
		got, err := transformAll(context.Background(), c.input, c.fn)
		if got != nil || err == nil {
			t.Fatal(got, err)
		}
	}
}
func TestTransformJoinsAllErrors(t *testing.T) {
	first := errors.New("first")
	second := errors.New("second")
	seen := make([]bool, 3)
	got, err := transformAll(context.Background(), []string{"a", "b", "c"}, func(ctx context.Context, input string) (string, error) {
		index := int(input[0] - 'a')
		seen[index] = true
		switch input {
		case "a":
			return "", first
		case "c":
			return "", second
		}
		return input, nil
	})
	if got != nil || !errors.Is(err, first) || !errors.Is(err, second) || !reflect.DeepEqual(seen, []bool{true, true, true}) {
		t.Fatal(got, err, seen)
	}
}
func TestTransformCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	got, err := transformAll(ctx, []string{"a"}, func(context.Context, string) (string, error) { called = true; return "", nil })
	if got != nil || !errors.Is(err, context.Canceled) || called {
		t.Fatal(got, err, called)
	}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	got, err = transformAll(ctx, []string{"a"}, func(context.Context, string) (string, error) { cancel(); return "a", nil })
	if got != nil || !errors.Is(err, context.Canceled) {
		t.Fatal(got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 정상 결과: [GO TEST]
	// 실패 결과 없음: true
	// 오류 확인: true
}
