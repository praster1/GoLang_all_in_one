package main

import "testing"

func TestEmbeddedMessage(t *testing.T) {
	t.Chdir(t.TempDir())
	if greeting() != "프로그램에 포함된 안내입니다." {
		t.Fatal("embedded content unavailable")
	}
}

func Example_main() {
	main()
	// Output:
	// 프로그램에 포함된 안내입니다.
}
