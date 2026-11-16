package postgres

import (
	"context"
	"errors"
	"io/fs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/example/tasks-api/internal/postgres/sqlc"
	"github.com/example/tasks-api/internal/task"
	"github.com/example/tasks-api/internal/user"
)

// PostgreSQL error code for unique_violation.
const uniqueViolation = "23505"

func mustSub(fsys fs.FS, dir string) fs.FS {
	sub, err := fs.Sub(fsys, dir)
	if err != nil {
		panic(err) // only possible with an invalid, hard-coded dir name
	}
	return sub
}

// ---- Users -----------------------------------------------------------------

// UserRepo implements user.Repository.
type UserRepo struct {
	q *sqlc.Queries
}

// Compile-time check that UserRepo satisfies the interface.
var _ user.Repository = (*UserRepo)(nil)

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{q: sqlc.New(pool)}
}

func (r *UserRepo) Create(ctx context.Context, email, name, hash string) (user.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{Email: email, Name: name, PasswordHash: hash})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
			return user.User{}, user.ErrEmailTaken
		}
		return user.User{}, err
	}
	return toUser(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return toUser(row), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	return toUser(row), nil
}

func toUser(row sqlc.User) user.User {
	return user.User{
		ID:           row.ID,
		Email:        row.Email,
		Name:         row.Name,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}

// ---- Tasks -----------------------------------------------------------------

// TaskRepo implements task.Repository.
type TaskRepo struct {
	q *sqlc.Queries
}

var _ task.Repository = (*TaskRepo)(nil)

func NewTaskRepo(pool *pgxpool.Pool) *TaskRepo {
	return &TaskRepo{q: sqlc.New(pool)}
}

func (r *TaskRepo) Create(ctx context.Context, p task.CreateParams) (task.Task, error) {
	row, err := r.q.CreateTask(ctx, sqlc.CreateTaskParams{
		UserID:      p.UserID,
		Title:       p.Title,
		Description: p.Description,
		Status:      string(p.Status),
		DueAt:       p.DueAt,
	})
	if err != nil {
		return task.Task{}, err
	}
	return toTask(row), nil
}

func (r *TaskRepo) Get(ctx context.Context, userID, id string) (task.Task, error) {
	row, err := r.q.GetTask(ctx, sqlc.GetTaskParams{ID: id, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return task.Task{}, task.ErrNotFound
	}
	if err != nil {
		return task.Task{}, err
	}
	return toTask(row), nil
}

func (r *TaskRepo) List(ctx context.Context, p task.ListParams) ([]task.Task, error) {
	arg := sqlc.ListTasksParams{
		UserID:   p.UserID,
		RowLimit: int32(p.Limit),
	}
	if p.Status != nil {
		s := string(*p.Status)
		arg.Status = &s
	}
	if p.After != nil {
		arg.CursorCreatedAt = &p.After.CreatedAt
		arg.CursorID = &p.After.ID
	}

	rows, err := r.q.ListTasks(ctx, arg)
	if err != nil {
		return nil, err
	}
	tasks := make([]task.Task, len(rows))
	for i, row := range rows {
		tasks[i] = toTask(row)
	}
	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, p task.UpdateParams) (task.Task, error) {
	arg := sqlc.UpdateTaskParams{
		ID:          p.ID,
		UserID:      p.UserID,
		Title:       p.Title,
		Description: p.Description,
		DueAt:       p.DueAt,
	}
	if p.Status != nil {
		s := string(*p.Status)
		arg.Status = &s
	}
	row, err := r.q.UpdateTask(ctx, arg)
	if errors.Is(err, pgx.ErrNoRows) {
		return task.Task{}, task.ErrNotFound
	}
	if err != nil {
		return task.Task{}, err
	}
	return toTask(row), nil
}

func (r *TaskRepo) Delete(ctx context.Context, userID, id string) error {
	n, err := r.q.DeleteTask(ctx, sqlc.DeleteTaskParams{ID: id, UserID: userID})
	if err != nil {
		return err
	}
	if n == 0 {
		return task.ErrNotFound
	}
	return nil
}

func toTask(row sqlc.Task) task.Task {
	return task.Task{
		ID:          row.ID,
		UserID:      row.UserID,
		Title:       row.Title,
		Description: row.Description,
		Status:      task.Status(row.Status),
		DueAt:       row.DueAt,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
