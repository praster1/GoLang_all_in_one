package main

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
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
func TestPoolDeadlineReleasesWaiter(t *testing.T) {
	db := testDB(t)
	held, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := countNotes(ctx, db); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if db.Stats().WaitCount < 1 {
		t.Fatal("query did not wait for pool")
	}
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	live, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if got, err := countNotes(live, db); err != nil || got != 2 {
		t.Fatalf("count=%d err=%v", got, err)
	}
}
func TestEmptyTableAndCanceledCall(t *testing.T) {
	db := testDB(t)
	if _, err := db.ExecContext(context.Background(), "DELETE FROM notes"); err != nil {
		t.Fatal(err)
	}
	if got, err := countNotes(context.Background(), db); err != nil || got != 0 {
		t.Fatalf("count=%d err=%v", got, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := countNotes(ctx, db); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
