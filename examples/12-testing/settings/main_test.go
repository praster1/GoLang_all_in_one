package main

import (
	"os"
	"testing"
)

func TestReadName(t *testing.T) {
	for _, tc := range []struct{ value, want string }{{"지수", "지수"}, {" 민수 ", "민수"}, {"", "방문자"}, {"  ", "방문자"}} {
		t.Run(tc.want+tc.value, func(t *testing.T) {
			got := readName(func(key string) string {
				if key != nameKey {
					t.Fatalf("key=%q; want %q", key, nameKey)
				}
				return tc.value
			})
			if got != tc.want {
				t.Errorf("got=%q; want %q", got, tc.want)
			}
		})
	}
}

func TestReadNameFromEnvironment(t *testing.T) {
	t.Setenv(nameKey, " 지수 ")
	if got := readName(os.Getenv); got != "지수" {
		t.Errorf("got=%q; want 지수", got)
	}
}

func Example_main() {
	main()
	// Output:
	// 이름: 학습자
}
