package main

import (
	"slices"
	"testing"
)

func TestCloneOrder(t *testing.T) {
	source := Order{ID: "A", Items: []Item{{Name: "book", Quantity: 2}, {Name: "pen", Quantity: 3}}}
	cloned := cloneOrder(source)
	if cloned.ID != "A" || !slices.Equal(cloned.Items, source.Items) {
		t.Fatalf("lost values: %+v", cloned)
	}
	cloned.ID = "B"
	cloned.Items[0].Quantity = 9
	cloned.Items[1].Name = "note"
	if source.ID != "A" || source.Items[0].Quantity != 2 || source.Items[1].Name != "pen" {
		t.Fatalf("source changed: %+v", source)
	}
	source.Items[1].Quantity = 10
	if cloned.Items[1].Quantity != 3 {
		t.Fatal("source change reached copy")
	}
	if cloneOrder(Order{}).Items != nil {
		t.Fatal("nil changed")
	}
	if cloneOrder(Order{Items: []Item{}}).Items == nil {
		t.Fatal("empty changed to nil")
	}
}

func Example_main() {
	main()
	// Output:
	// 원본: A-1, book 2개
	// 복사본: A-2, book 9개
}
