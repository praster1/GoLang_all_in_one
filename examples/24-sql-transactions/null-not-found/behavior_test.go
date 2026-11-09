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
	return db
}
func TestNullableValuesAndMissingRow(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if err := createSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		id    int64
		valid bool
		title string
	}{{1, true, "Go"}, {2, false, ""}, {3, true, ""}} {
		got, err := findNote(ctx, db, tc.id)
		if err != nil {
			t.Fatal(err)
		}
		if got.ID != tc.id || got.Title.Valid != tc.valid || got.Title.String != tc.title {
			t.Fatalf("id=%d got=%+v", tc.id, got)
		}
	}
	if _, err := findNote(ctx, db, 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err=%v", err)
	}
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := findNote(ctx, db, 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
