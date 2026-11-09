package main

import "testing"

func TestLoadConfig(t *testing.T) {
	for _, tc := range []struct {
		value  string
		exists bool
		want   int
		bad    bool
	}{
		{"", false, 8080, false}, {"1", true, 1, false}, {"65535", true, 65535, false},
		{"", true, 0, true}, {"0", true, 0, true}, {"65536", true, 0, true}, {"oops", true, 0, true},
	} {
		got, err := loadConfig(func(string) (string, bool) { return tc.value, tc.exists })
		if (err != nil) != tc.bad || got.Port != tc.want {
			t.Fatalf("got=%v err=%v", got, err)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// 실행 설정: 9090
}
