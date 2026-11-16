// Package task is the "tasks" feature: domain model, business rules
// (Service), storage contract (Repository) and HTTP transport (Handler).
package task

import (
	"context"
	"errors"
	"time"
)

// Status is the lifecycle state of a task.
type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Valid reports whether s is a known status.
func (s Status) Valid() bool {
	switch s {
	case StatusTodo, StatusInProgress, StatusDone:
		return true
	}
	return false
}

// Task is the domain model. It has no JSON or SQL tags on purpose:
// the HTTP layer and the database layer each map it to their own shape.
type Task struct {
	ID          string
	UserID      string
	Title       string
	Description string
	Status      Status
	DueAt       *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ErrNotFound is returned when a task does not exist or belongs to another user.
var ErrNotFound = errors.New("task not found")

// Cursor marks a position in a list ordered by (CreatedAt DESC, ID DESC).
type Cursor struct {
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"id"`
}

type CreateParams struct {
	UserID      string
	Title       string
	Description string
	Status      Status
	DueAt       *time.Time
}

// UpdateParams uses pointers: nil means "do not change this field".
type UpdateParams struct {
	ID          string
	UserID      string
	Title       *string
	Description *string
	Status      *Status
	DueAt       *time.Time
}

type ListParams struct {
	UserID string
	Status *Status
	After  *Cursor
	Limit  int
}

// Repository is the storage contract the service depends on.
// It is defined here, by the consumer, so the service does not know
// whether data lives in PostgreSQL, memory or anywhere else.
type Repository interface {
	Create(ctx context.Context, p CreateParams) (Task, error)
	Get(ctx context.Context, userID, id string) (Task, error)
	List(ctx context.Context, p ListParams) ([]Task, error)
	Update(ctx context.Context, p UpdateParams) (Task, error)
	Delete(ctx context.Context, userID, id string) error
}
