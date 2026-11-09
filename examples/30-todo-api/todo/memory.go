package todo

import (
	"context"
	"math"
	"slices"
	"sync"
)

type Memory struct {
	mu     sync.Mutex
	nextID int64
	tasks  map[int64]Task
}

func NewMemory() *Memory { return &Memory{nextID: 1, tasks: make(map[int64]Task)} }

var _ Repository = (*Memory)(nil)

func (m *Memory) Create(ctx context.Context, title string) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	title, err := NormalizeTitle(title)
	if err != nil {
		return Task{}, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if m.nextID <= 0 {
		return Task{}, ErrIDExhausted
	}
	task := Task{ID: m.nextID, Title: title}
	if m.nextID == math.MaxInt64 {
		m.nextID = 0
	} else {
		m.nextID++
	}
	m.tasks[task.ID] = task
	return task, nil
}
func (m *Memory) Get(ctx context.Context, id int64) (Task, error) {
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if id <= 0 {
		return Task{}, ErrInvalidID
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	task, ok := m.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return task, nil
}
func (m *Memory) List(ctx context.Context, after int64, limit int) ([]Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := validatePage(after, limit); err != nil {
		return nil, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(m.tasks))
	for id := range m.tasks {
		if id > after {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	if len(ids) > limit {
		ids = ids[:limit]
	}
	result := make([]Task, 0, len(ids))
	for _, id := range ids {
		result = append(result, m.tasks[id])
	}
	return result, nil
}
func (m *Memory) Replace(ctx context.Context, id int64, title string, done bool) (Task, error) {
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
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return Task{}, err
	}
	if _, ok := m.tasks[id]; !ok {
		return Task{}, ErrNotFound
	}
	task := Task{ID: id, Title: title, Done: done}
	m.tasks[id] = task
	return task, nil
}
func (m *Memory) Delete(ctx context.Context, id int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if id <= 0 {
		return ErrInvalidID
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, ok := m.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(m.tasks, id)
	return nil
}
