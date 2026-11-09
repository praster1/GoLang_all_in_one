package main

import "testing"

func TestAddInPlaceAndRedirect(t *testing.T) {
	value := 10
	if result := added(value, 3); result != 13 || value != 10 {
		t.Fatalf("added: value=%d", value)
	}
	if !addInPlace(&value, -3) || value != 7 {
		t.Fatalf("negative amount: %d", value)
	}
	if !addInPlace(&value, 0) || value != 7 {
		t.Fatalf("zero amount: %d", value)
	}
	redirect(&value)
	if value != 7 {
		t.Fatalf("redirect changed original: %d", value)
	}
	if addInPlace(nil, 1) {
		t.Fatal("nil accepted")
	}
}

func Example_main() {
	main()
	// Output:
	// 값으로 전달: 10 13
	// 주소로 접근: 15 true
	// 포인터 재대입 뒤: 15
	// nil 전달: false
}
