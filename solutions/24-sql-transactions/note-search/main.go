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

type Note struct {
	ID    int64
	Title string
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes(id INTEGER PRIMARY KEY, title TEXT NOT NULL UNIQUE);
 INSERT INTO notes VALUES(1, 'Go''s SQL'),(2,'SQL')`)
	return err
}
func findByTitle(ctx context.Context, db *sql.DB, title string) (Note, error) {
	var note Note
	err := db.QueryRowContext(ctx, "SELECT id,title FROM notes WHERE title = ?", title).Scan(&note.ID, &note.Title)
	if err != nil {
		return Note{}, fmt.Errorf("find title: %w", err)
	}
	return note, nil
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
	note, err := findByTitle(ctx, db, "Go's SQL")
	if err != nil {
		return err
	}
	fmt.Printf("검색: %d, %s\n", note.ID, note.Title)
	_, missing := findByTitle(ctx, db, "' OR 1=1 --")
	fmt.Println("문자 그대로 검색:", errors.Is(missing, sql.ErrNoRows))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
