package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

type SQLite struct{ db *sql.DB }

var _ Repository = (*SQLite)(nil)

func OpenSQLite(ctx context.Context, dsn string) (*SQLite, error) {
	if dsn == "" {
		return nil, errors.New("SQLite 경로가 필요합니다")
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	// :memory: is per physical connection. The tutorial keeps a single connection.
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(err, db.Close())
	}
	_, err = db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS todos (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL CHECK(length(title) BETWEEN 1 AND 80),
        done INTEGER NOT NULL DEFAULT 0 CHECK(done IN (0,1))
    )`)
	if err != nil {
		return nil, errors.Join(err, db.Close())
	}
	return &SQLite{db: db}, nil
}
func (s *SQLite) Close() error                   { return s.db.Close() }
func (s *SQLite) Ping(ctx context.Context) error { return s.db.PingContext(ctx) }
func scanTask(row interface{ Scan(...any) error }) (Task, error) {
	var task Task
	if err := row.Scan(&task.ID, &task.Title, &task.Done); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Task{}, ErrNotFound
		}
		return Task{}, err
	}
	return task, nil
}
func (s *SQLite) Create(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	row := s.db.QueryRowContext(ctx, "INSERT INTO todos(title) VALUES (?) RETURNING id,title,done", title)
	task, err := scanTask(row)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return task, nil
}
func (s *SQLite) Get(ctx context.Context, id int64) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	return scanTask(s.db.QueryRowContext(ctx, "SELECT id,title,done FROM todos WHERE id=?", id))
}
func (s *SQLite) List(ctx context.Context, after int64, limit int) (result []Task, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePage(after, limit); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, "SELECT id,title,done FROM todos WHERE id>? ORDER BY id LIMIT ?", after, limit)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, rows.Close()) }()
	result = make([]Task, 0)
	for rows.Next() {
		task, scanErr := scanTask(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, task)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
func (s *SQLite) Replace(ctx context.Context, id int64, title string, done bool) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	return scanTask(s.db.QueryRowContext(ctx, "UPDATE todos SET title=?,done=? WHERE id=? RETURNING id,title,done", title, done, id))
}
func (s *SQLite) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id <= 0 {
		return ErrInvalidID
	}
	result, err := s.db.ExecContext(ctx, "DELETE FROM todos WHERE id=?", id)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
