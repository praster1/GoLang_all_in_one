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
func TestTransferCommitAndRollback(t *testing.T) {
	for _, tc := range []struct {
		name          string
		from, to      int64
		amount        int
		wantErr       error
		first, second int
	}{
		{"commit", 1, 2, 30, nil, 70, 50}, {"all points", 1, 2, 100, nil, 0, 120},
		{"insufficient", 1, 2, 101, ErrSource, 100, 20}, {"missing source", 99, 2, 10, ErrSource, 100, 20},
		{"missing destination", 1, 99, 10, ErrDestination, 100, 20}, {"zero", 1, 2, 0, ErrInput, 100, 20},
		{"negative", 1, 2, -1, ErrInput, 100, 20}, {"same", 1, 1, 1, ErrInput, 100, 20},
		{"large", 1, 2, 1001, ErrInput, 100, 20},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := testDB(t)
			ctx := context.Background()
			if err := createSchema(ctx, db); err != nil {
				t.Fatal(err)
			}
			if err := transfer(ctx, db, tc.from, tc.to, tc.amount); !errors.Is(err, tc.wantErr) {
				t.Fatalf("err=%v want=%v", err, tc.wantErr)
			}
			first, err := points(ctx, db, 1)
			if err != nil {
				t.Fatal(err)
			}
			second, err := points(ctx, db, 2)
			if err != nil {
				t.Fatal(err)
			}
			if first != tc.first || second != tc.second || first+second != 120 {
				t.Fatalf("points=%d,%d", first, second)
			}
		})
	}
}
func TestCanceledTransferDoesNotChangePoints(t *testing.T) {
	db := testDB(t)
	if err := createSchema(context.Background(), db); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := transfer(ctx, db, 1, 2, 30); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	got, err := points(context.Background(), db, 1)
	if err != nil || got != 100 {
		t.Fatalf("got=%d err=%v", got, err)
	}
}
