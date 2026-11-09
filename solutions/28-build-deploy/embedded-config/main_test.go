package main

import "testing"

func TestConfig(t *testing.T) {
	t.Chdir(t.TempDir())
	cfg, err := parseConfig(defaults)
	if err != nil || cfg.Name != "demo" || cfg.Port != 8080 {
		t.Fatal("embedded defaults")
	}
	for _, text := range []string{`{}`, `null`, `{"name":"x","port":0}`, `{"name":"x","port":80,"extra":1}`, `{"name":"x","port":80} {}`} {
		if _, err := parseConfig(text); err == nil {
			t.Fatalf("accepted %s", text)
		}
	}
}

func Example_main() {
	main()
	// Output:
	// demo:8080
}
