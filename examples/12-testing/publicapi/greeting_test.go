package publicapi_test

import (
	"fmt"
	"testing"

	"github.com/praster1/GoLang_all_in_one/examples/12-testing/publicapi"
)

func TestGreeting(t *testing.T) {
	for _, tc := range []struct{ name, want string }{{" 지수 ", "지수님, 안녕하세요."}, {"", "방문자님, 안녕하세요."}, {" \n", "방문자님, 안녕하세요."}} {
		if got := publicapi.Greeting(tc.name); got != tc.want {
			t.Errorf("Greeting(%q)=%q; want %q", tc.name, got, tc.want)
		}
	}
}

func ExampleGreeting() {
	fmt.Println(publicapi.Greeting("지수"))
	// Output:
	// 지수님, 안녕하세요.
}

func ExampleGreeting_empty() {
	fmt.Println(publicapi.Greeting(" "))
	// Output:
	// 방문자님, 안녕하세요.
}
