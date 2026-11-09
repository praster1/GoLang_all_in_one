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
	return db
}
func TestPoolWaitDeadlineAndReuse(t *testing.T) {
	db := testDB(t)
	held, err := db.Conn(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer held.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if _, err := readValue(ctx, db); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
	if db.Stats().WaitCount < 1 {
		t.Fatal("expected a connection pool wait")
	}
	if err := held.Close(); err != nil {
		t.Fatal(err)
	}
	check, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if got, err := readValue(check, db); err != nil || got != 1 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
func TestAlreadyCanceledQuery(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := readValue(ctx, db); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
