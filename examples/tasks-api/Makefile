.PHONY: run test test-integration lint sqlc db-up db-down docker-up docker-down

# Load .env if it exists (copy from .env.example).
-include .env
export

run: ## Run the API locally (needs a database: make db-up)
	go run ./cmd/api

test: ## Unit tests with the race detector
	go test -race ./...

test-integration: ## Tests against the local PostgreSQL
	TEST_DATABASE_URL=postgres://tasks:tasks@localhost:$${DB_HOST_PORT:-5433}/tasks?sslmode=disable go test -race -count=1 ./...

lint: ## Static checks
	go vet ./...
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

sqlc: ## Regenerate type-safe query code after editing internal/postgres/queries
	docker run --rm -v "$(CURDIR)":/src -w /src sqlc/sqlc:1.31.1 generate

db-up: ## Start only PostgreSQL
	docker compose up -d db

db-down:
	docker compose down

docker-up: ## Build and run API + PostgreSQL
	docker compose up --build -d

docker-down:
	docker compose down
