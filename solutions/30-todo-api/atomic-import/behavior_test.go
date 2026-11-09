package main

import (
	"context"
	"errors"
	"testing"
)

func TestImportAtomicity(t *testing.T) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := importTasks(ctx, db, []string{"existing"}); err != nil {
		t.Fatal(err)
	}
	if err := importTasks(ctx, db, []string{"new", "existing"}); err == nil {
		t.Fatal("duplicate accepted")
	}
	n, err := count(ctx, db)
	if err != nil || n != 1 {
		t.Fatalf("partial commit n=%d err=%v", n, err)
	}
	if err := importTasks(ctx, db, []string{"valid", " "}); err == nil {
		t.Fatal("invalid title accepted")
	}
	n, err = count(ctx, db)
	if err != nil || n != 1 {
		t.Fatal("invalid batch changed state")
	}
	if err := importTasks(ctx, db, []string{" Go ", "SQL"}); err != nil {
		t.Fatal(err)
	}
	n, err = count(ctx, db)
	if err != nil || n != 3 {
		t.Fatal("valid batch not stored")
	}
	for _, titles := range [][]string{nil, make([]string, 101)} {
		if err := importTasks(ctx, db, titles); err == nil {
			t.Fatal("accepted batch size")
		}
	}
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	if err := importTasks(canceled, db, []string{"canceled"}); !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v", err)
	}
}
