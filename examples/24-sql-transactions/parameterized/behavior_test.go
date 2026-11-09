package main

import (
	"context"
	"database/sql"
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
func TestValuesRemainData(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if err := createSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	for index, title := range []string{"Go's SQL", "x'); DROP TABLE notes; --", "", "한글"} {
		id, err := addNote(ctx, db, title)
		if err != nil {
			t.Fatal(err)
		}
		if id != int64(index+1) {
			t.Fatalf("id=%d", id)
		}
		var got string
		if err := db.QueryRowContext(ctx, "SELECT title FROM notes WHERE id = ?", id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != title {
			t.Fatalf("got=%q want=%q", got, title)
		}
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil || count != 4 {
		t.Fatalf("count=%d err=%v", count, err)
	}
}
