package task

import (
	"context"
	"fmt"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Service holds the business rules for tasks. It knows nothing about HTTP.
type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, p CreateParams) (Task, error) {
	if p.Status == "" {
		p.Status = StatusTodo
	}
	t, err := s.repo.Create(ctx, p)
	if err != nil {
		return Task{}, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}

func (s *Service) Get(ctx context.Context, userID, id string) (Task, error) {
	return s.repo.Get(ctx, userID, id)
}

// Page is one page of a cursor-paginated list.
type Page struct {
	Tasks []Task
	Next  *Cursor // nil when there are no more results
}

// List returns one page. It asks the repository for limit+1 rows: if the
// extra row exists there is another page, and no COUNT(*) is needed.
func (s *Service) List(ctx context.Context, p ListParams) (Page, error) {
	p.Limit = clampLimit(p.Limit)
	limit := p.Limit
	p.Limit++

	tasks, err := s.repo.List(ctx, p)
	if err != nil {
		return Page{}, fmt.Errorf("list tasks: %w", err)
	}

	page := Page{Tasks: tasks}
	if len(tasks) > limit {
		page.Tasks = tasks[:limit]
		last := page.Tasks[limit-1]
		page.Next = &Cursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}
	return page, nil
}

func (s *Service) Update(ctx context.Context, p UpdateParams) (Task, error) {
	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

// clampLimit applies the default and the server-side maximum.
// Never trust a page size sent by the client.
func clampLimit(n int) int {
	if n <= 0 {
		return DefaultPageSize
	}
	return min(n, MaxPageSize)
}
