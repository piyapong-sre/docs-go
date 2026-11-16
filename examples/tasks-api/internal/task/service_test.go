package task_test

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/example/tasks-api/internal/task"
)

// memRepo is an in-memory task.Repository used only in tests.
// Because the service depends on an interface, we can test business
// rules without a database.
type memRepo struct {
	mu    sync.Mutex
	tasks []task.Task
	seq   int
	clock time.Time
}

func (m *memRepo) Create(_ context.Context, p task.CreateParams) (task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	m.clock = m.clock.Add(time.Second)
	t := task.Task{
		ID:          fmt.Sprintf("00000000-0000-7000-8000-%012d", m.seq),
		UserID:      p.UserID,
		Title:       p.Title,
		Description: p.Description,
		Status:      p.Status,
		DueAt:       p.DueAt,
		CreatedAt:   m.clock,
		UpdatedAt:   m.clock,
	}
	m.tasks = append(m.tasks, t)
	return t, nil
}

func (m *memRepo) Get(_ context.Context, userID, id string) (task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.tasks {
		if t.ID == id && t.UserID == userID {
			return t, nil
		}
	}
	return task.Task{}, task.ErrNotFound
}

func (m *memRepo) List(_ context.Context, p task.ListParams) ([]task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []task.Task
	for _, t := range m.tasks {
		if t.UserID != p.UserID || (p.Status != nil && t.Status != *p.Status) {
			continue
		}
		if p.After != nil && !before(t, *p.After) {
			continue
		}
		out = append(out, t)
	}
	// newest first, same as ORDER BY created_at DESC, id DESC
	slices.SortFunc(out, func(a, b task.Task) int {
		return cmp.Or(b.CreatedAt.Compare(a.CreatedAt), cmp.Compare(b.ID, a.ID))
	})
	return out[:min(len(out), p.Limit)], nil
}

// before reports whether t comes after cursor c in DESC order.
func before(t task.Task, c task.Cursor) bool {
	if !t.CreatedAt.Equal(c.CreatedAt) {
		return t.CreatedAt.Before(c.CreatedAt)
	}
	return t.ID < c.ID
}

func (m *memRepo) Update(_ context.Context, p task.UpdateParams) (task.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID != p.ID || t.UserID != p.UserID {
			continue
		}
		if p.Title != nil {
			t.Title = *p.Title
		}
		if p.Description != nil {
			t.Description = *p.Description
		}
		if p.Status != nil {
			t.Status = *p.Status
		}
		if p.DueAt != nil {
			t.DueAt = p.DueAt
		}
		m.tasks[i] = t
		return t, nil
	}
	return task.Task{}, task.ErrNotFound
}

func (m *memRepo) Delete(_ context.Context, userID, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, t := range m.tasks {
		if t.ID == id && t.UserID == userID {
			m.tasks = slices.Delete(m.tasks, i, i+1)
			return nil
		}
	}
	return task.ErrNotFound
}

func TestServiceCreateDefaultsStatus(t *testing.T) {
	svc := task.NewService(&memRepo{})
	got, err := svc.Create(t.Context(), task.CreateParams{UserID: "u1", Title: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != task.StatusTodo {
		t.Errorf("status = %q, want %q", got.Status, task.StatusTodo)
	}
}

func TestServiceListPaginates(t *testing.T) {
	ctx := t.Context()
	svc := task.NewService(&memRepo{})
	for i := range 5 {
		if _, err := svc.Create(ctx, task.CreateParams{UserID: "u1", Title: fmt.Sprint("t", i)}); err != nil {
			t.Fatal(err)
		}
	}
	// Another user's task must never appear.
	if _, err := svc.Create(ctx, task.CreateParams{UserID: "u2", Title: "other"}); err != nil {
		t.Fatal(err)
	}

	var titles []string
	var after *task.Cursor
	pages := 0
	for {
		page, err := svc.List(ctx, task.ListParams{UserID: "u1", Limit: 2, After: after})
		if err != nil {
			t.Fatal(err)
		}
		pages++
		for _, tk := range page.Tasks {
			titles = append(titles, tk.Title)
		}
		if page.Next == nil {
			break
		}
		after = page.Next
	}

	want := []string{"t4", "t3", "t2", "t1", "t0"}
	if !slices.Equal(titles, want) {
		t.Errorf("titles = %v, want %v", titles, want)
	}
	if pages != 3 {
		t.Errorf("pages = %d, want 3", pages)
	}
}

func TestCursorRoundTrip(t *testing.T) {
	c := task.Cursor{CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 6, time.UTC), ID: "abc"}
	got, err := task.DecodeCursor(task.EncodeCursor(c))
	if err != nil {
		t.Fatal(err)
	}
	if !got.CreatedAt.Equal(c.CreatedAt) || got.ID != c.ID {
		t.Errorf("got %+v, want %+v", got, c)
	}
	if _, err := task.DecodeCursor("not-a-cursor"); err == nil {
		t.Error("expected error for garbage cursor")
	}
}
