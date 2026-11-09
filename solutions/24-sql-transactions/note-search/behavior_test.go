package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := openDB(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	return db
}
func TestSearchExactTitle(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	note, err := findByTitle(ctx, db, "Go's SQL")
	if err != nil || note.ID != 1 || note.Title != "Go's SQL" {
		t.Fatalf("note=%+v err=%v", note, err)
	}
	for _, title := range []string{"", "missing", "' OR 1=1 --", "SQL' --"} {
		if _, err := findByTitle(ctx, db, title); !errors.Is(err, sql.ErrNoRows) {
			t.Fatalf("title=%q err=%v", title, err)
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := findByTitle(ctx, db, "SQL"); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
