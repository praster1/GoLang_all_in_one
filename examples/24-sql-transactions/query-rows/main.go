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

type Lesson struct {
	ID      int64
	Title   string
	Minutes int
}

func createSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE lessons (id INTEGER PRIMARY KEY, title TEXT NOT NULL, minutes INTEGER NOT NULL);
        INSERT INTO lessons VALUES (2, '복습', 15), (1, 'Go', 25), (3, 'SQL', 40)`)
	return err
}

func listLessons(ctx context.Context, db *sql.DB, minimum int) (lessons []Lesson, err error) {
	rows, err := db.QueryContext(ctx, "SELECT id, title, minutes FROM lessons WHERE minutes >= ? ORDER BY id", minimum)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	lessons = make([]Lesson, 0)
	for rows.Next() {
		var lesson Lesson
		if err := rows.Scan(&lesson.ID, &lesson.Title, &lesson.Minutes); err != nil {
			return nil, err
		}
		lessons = append(lessons, lesson)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return lessons, nil
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
	lessons, err := listLessons(ctx, db, 20)
	if err != nil {
		return err
	}
	for _, lesson := range lessons {
		fmt.Printf("%d: %s, %d분\n", lesson.ID, lesson.Title, lesson.Minutes)
	}
	return nil
}
func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
