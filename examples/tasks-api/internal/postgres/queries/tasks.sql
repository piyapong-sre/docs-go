-- name: CreateTask :one
INSERT INTO tasks (user_id, title, description, status, due_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetTask :one
-- Always filter by user_id so a user can never read another user's task.
SELECT * FROM tasks
WHERE id = $1 AND user_id = $2;

-- name: ListTasks :many
-- Keyset pagination: rows strictly "after" the cursor (created_at, id), newest first.
SELECT * FROM tasks
WHERE user_id = sqlc.arg(user_id)
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status))
  AND (
        sqlc.narg(cursor_created_at)::timestamptz IS NULL
        OR (created_at, id) < (sqlc.narg(cursor_created_at), sqlc.narg(cursor_id)::uuid)
      )
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(row_limit);

-- name: UpdateTask :one
-- NULL arguments mean "leave the column unchanged" (PATCH semantics).
UPDATE tasks
SET title       = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    status      = COALESCE(sqlc.narg(status), status),
    due_at      = COALESCE(sqlc.narg(due_at), due_at),
    updated_at  = now()
WHERE id = sqlc.arg(id) AND user_id = sqlc.arg(user_id)
RETURNING *;

-- name: DeleteTask :execrows
DELETE FROM tasks
WHERE id = $1 AND user_id = $2;
