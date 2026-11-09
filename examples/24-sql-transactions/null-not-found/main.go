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
	Title sql.NullString
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE notes (id INTEGER PRIMARY KEY, title TEXT);
        INSERT INTO notes VALUES (1, 'Go'), (2, NULL), (3, '')`)
	return err
}

func findNote(ctx context.Context, db *sql.DB, id int64) (Note, error) {
	var note Note
	err := db.QueryRowContext(ctx, "SELECT id, title FROM notes WHERE id = ?", id).Scan(&note.ID, &note.Title)
	if err != nil {
		return Note{}, fmt.Errorf("find note: %w", err)
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
	for _, id := range []int64{1, 2, 3} {
		note, err := findNote(ctx, db, id)
		if err != nil {
			return err
		}
		if note.Title.Valid {
			fmt.Printf("%d: %q\n", note.ID, note.Title.String)
		} else {
			fmt.Printf("%d: NULL\n", note.ID)
		}
	}
	_, lookupErr := findNote(ctx, db, 99)
	fmt.Println("없음:", errors.Is(lookupErr, sql.ErrNoRows))
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
