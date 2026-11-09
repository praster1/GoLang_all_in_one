package order

import "testing"

func TestNew(t *testing.T) {
	for _, tc := range []struct {
		quantity, price, total int
		ok                     bool
	}{{0, 0, 0, true}, {3, 2000, 6000, true}, {1000, 1000000, 1000000000, true}, {-1, 1, 0, false}, {1, -1, 0, false}, {1001, 1, 0, false}, {1, 1000001, 0, false}} {
		got, ok := New(tc.quantity, tc.price)
		if ok != tc.ok || got.Total() != tc.total {
			t.Fatalf("%+v: %+v %t", tc, got, ok)
		}
	}
}
