package main

import "testing"

func requireMinutes(t *testing.T, text string, want int) {
	t.Helper()
	got, err := parseClock(text)
	if err != nil {
		t.Fatalf("parseClock(%q): %v", text, err)
	}
	if got != want {
		t.Errorf("parseClock(%q) = %d; want %d", text, got, want)
	}
}

func TestParseClock(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		for _, tc := range []struct {
			name, text string
			want       int
		}{
			{"midnight", "00:00", 0}, {"morning", "09:30", 570}, {"last_minute", "23:59", 1439},
		} {
			t.Run(tc.name, func(t *testing.T) { requireMinutes(t, tc.text, tc.want) })
		}
	})
	t.Run("invalid", func(t *testing.T) {
		for _, tc := range []struct{ name, text string }{
			{"short", "9:30"}, {"separator", "09-30"}, {"hour", "24:00"},
			{"minute", "12:60"}, {"sign", "+1:00"}, {"space", " 9:30"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				got, err := parseClock(tc.text)
				if err == nil || got != 0 {
					t.Fatalf("parseClock(%q) = %d, %v; want rejection", tc.text, got, err)
				}
			})
		}
	})
}

func Example_main() {
	main()
	// Output:
	// 경과 분: 570
}
