package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
)

func openDB(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return db, nil
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT NOT NULL)")
	return err
}

func addNote(ctx context.Context, db *sql.DB, title string) (int64, error) {
	result, err := db.ExecContext(ctx, "INSERT INTO notes(title) VALUES (?)", title)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func run() (err error) {
	ctx := context.Background()
	db, err := openDB(ctx)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	if err := createSchema(ctx, db); err != nil {
		return err
	}
	for _, title := range []string{"Go's SQL", "x'); DROP TABLE notes; --"} {
		id, err := addNote(ctx, db, title)
		if err != nil {
			return err
		}
		fmt.Printf("저장: %d, %s\n", id, title)
	}
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM notes").Scan(&count); err != nil {
		return err
	}
	fmt.Println("행 수:", count)
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
