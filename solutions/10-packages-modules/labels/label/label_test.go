package label

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	for _, tc := range []struct {
		in, want string
		ok       bool
	}{{" GO ", "go", true}, {" 한글 ", "한글", true}, {" ", "", false}, {"a\nb", "", false}, {string([]byte{255}), "", false}, {strings.Repeat("가", 20), strings.Repeat("가", 20), true}, {strings.Repeat("가", 21), "", false}} {
		got, ok := Normalize(tc.in)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("%q: %q %t", tc.in, got, ok)
		}
	}
}
