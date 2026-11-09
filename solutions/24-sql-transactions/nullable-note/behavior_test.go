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
func TestThreeTitleStates(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	for _, tc := range []struct {
		id   int64
		want string
	}{{1, "Go"}, {2, "(미정)"}, {3, "(빈 문자열)"}} {
		got, err := noteLabel(ctx, db, tc.id)
		if err != nil || got != tc.want {
			t.Fatalf("id=%d got=%q err=%v", tc.id, got, err)
		}
	}
	if _, err := noteLabel(ctx, db, 99); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("err=%v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := noteLabel(ctx, db, 1); err == nil || errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("closed database error=%v", err)
	}
}
