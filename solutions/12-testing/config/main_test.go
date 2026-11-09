package main

import (
	"errors"
	"os"
	"testing"
)

func TestLoadLimit(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		present    bool
		want       int
		bad        bool
	}{
		{"unset", "", false, 20, false}, {"empty", "", true, 0, true}, {"minimum", "1", true, 1, false},
		{"leading_zero", "0030", true, 30, false}, {"maximum", "100", true, 100, false},
		{"zero", "0", true, 0, true}, {"over", "101", true, 0, true}, {"sign", "+1", true, 0, true},
		{"space", " 1", true, 0, true}, {"overflow", "999999999999999999999", true, 0, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := loadLimit(func(key string) (string, bool) {
				if key != limitKey {
					t.Fatalf("key=%q", key)
				}
				return tc.text, tc.present
			})
			if (err != nil) != tc.bad || (tc.bad && !errors.Is(err, errLimit)) || got != tc.want {
				t.Fatalf("got=%d err=%v; want %d bad=%v", got, err, tc.want, tc.bad)
			}
		})
	}
}

func TestLoadLimitEnvironment(t *testing.T) {
	t.Setenv(limitKey, "42")
	got, err := loadLimit(os.LookupEnv)
	if err != nil || got != 42 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}

func Example_main() {
	main()
	// Output:
	// 한도: 30
}
