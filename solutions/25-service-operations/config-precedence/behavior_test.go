package main

import "testing"

func TestPrecedence(t *testing.T) {
	env := func(string) (string, bool) { return "30", true }
	for _, tc := range []struct {
		args []string
		want int
	}{{nil, 30}, {[]string{"-limit=1"}, 1}, {[]string{"-limit=100"}, 100}, {[]string{"-limit=40"}, 40}} {
		got, err := limitConfig(tc.args, env)
		if err != nil || got != tc.want {
			t.Fatalf("got=%d error=%v", got, err)
		}
	}
	for _, args := range [][]string{{"-limit=0"}, {"-limit=-1"}, {"-limit=101"}, {"-limit=x"}, {"-unknown"}, {"extra"}} {
		if _, err := limitConfig(args, env); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
	if _, err := limitConfig([]string{"-limit=20"}, func(string) (string, bool) { return "bad", true }); err == nil {
		t.Fatal("invalid environment silently replaced")
	}
}
