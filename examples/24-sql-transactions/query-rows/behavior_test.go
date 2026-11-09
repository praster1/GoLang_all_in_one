package main

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
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
func TestListFilterAndOrder(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if err := createSchema(ctx, db); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		minimum int
		ids     []int64
	}{{0, []int64{1, 2, 3}}, {20, []int64{1, 3}}, {40, []int64{3}}, {41, []int64{}}} {
		got, err := listLessons(ctx, db, tc.minimum)
		if err != nil {
			t.Fatal(err)
		}
		ids := make([]int64, 0, len(got))
		for _, lesson := range got {
			ids = append(ids, lesson.ID)
		}
		if !reflect.DeepEqual(ids, tc.ids) {
			t.Fatalf("minimum=%d ids=%v want=%v", tc.minimum, ids, tc.ids)
		}
	}
}
func TestScanFailureReleasesConnection(t *testing.T) {
	db := testDB(t)
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `CREATE TABLE lessons(id INTEGER, title TEXT, minutes TEXT); INSERT INTO lessons VALUES(1,'Go','not-an-integer')`); err != nil {
		t.Fatal(err)
	}
	if _, err := listLessons(ctx, db, 0); err == nil {
		t.Fatal("expected scan error")
	}
	check, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	if err := db.PingContext(check); err != nil {
		t.Fatalf("connection not released: %v", err)
	}
}
func TestCanceledQuery(t *testing.T) {
	db := testDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := listLessons(ctx, db, 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
}
