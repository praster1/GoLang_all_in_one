package main

import "testing"

func TestLabelNilContract(t *testing.T) {
	var pointer *User
	var value Labeler = pointer
	if value == nil || value.Label() != "사용자 없음" {
		t.Fatal("typed nil contract")
	}
	if (&User{Name: "Mina"}).Label() != "사용자: Mina" {
		t.Fatal("name lost")
	}
}

func Example_main() {
	main()
	// Output:
	// 빈 인터페이스: true
	// 포인터: true
	// 포인터를 담은 인터페이스: false
	// 동적 타입: *main.User
	// 사용자 없음
}
