package counter_test

import (
	"github.com/praster1/GoLang_all_in_one/examples/10-packages-modules/private-state/counter"
	"testing"
)

func TestPublicContract(t *testing.T) {
	var c counter.Counter
	if !c.Add(100) || c.Value() != 100 {
		t.Fatal("100 boundary")
	}
	if c.Add(1) || c.Add(-1) || c.Value() != 100 {
		t.Fatal("failed change mutated state")
	}
	var missing *counter.Counter
	if missing.Add(1) {
		t.Fatal("nil accepted")
	}
}
