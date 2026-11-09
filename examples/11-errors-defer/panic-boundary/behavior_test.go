package main

import (
	"strings"
	"testing"
)

func TestKnownPanic(t *testing.T) {
	if err := invoke(func() {}); err != nil {
		t.Fatal(err)
	}
	if err := invoke(func() { panic(stopped{"stop"}) }); err == nil || !strings.Contains(err.Error(), "stop") {
		t.Fatal(err)
	}
}
func TestUnknownPanicPropagates(t *testing.T) {
	token := &struct{}{}
	defer func() {
		if got := recover(); got != token {
			t.Fatalf("got=%v want original panic", got)
		}
	}()
	_ = invoke(func() { panic(token) })
	t.Fatal("panic swallowed")
}
