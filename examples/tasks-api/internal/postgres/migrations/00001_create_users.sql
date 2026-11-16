-- +goose Up
CREATE TABLE users (
    id            uuid        PRIMARY KEY DEFAULT uuidv7(),
    email         text        NOT NULL,
    name          text        NOT NULL,
    password_hash text        NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Emails are stored lower-cased by the application; the index enforces uniqueness.
CREATE UNIQUE INDEX users_email_key ON users (email);

-- +goose Down
DROP TABLE users;
