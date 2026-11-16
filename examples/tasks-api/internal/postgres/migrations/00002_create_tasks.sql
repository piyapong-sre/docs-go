-- +goose Up
CREATE TABLE tasks (
    id          uuid        PRIMARY KEY DEFAULT uuidv7(),
    user_id     uuid        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title       text        NOT NULL,
    description text        NOT NULL DEFAULT '',
    status      text        NOT NULL DEFAULT 'todo'
                            CHECK (status IN ('todo', 'in_progress', 'done')),
    due_at      timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

-- Supports "list my tasks, newest first" with keyset (cursor) pagination.
CREATE INDEX tasks_user_created_idx ON tasks (user_id, created_at DESC, id DESC);

-- +goose Down
DROP TABLE tasks;
