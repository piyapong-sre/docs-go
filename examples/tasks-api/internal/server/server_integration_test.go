package server_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/example/tasks-api/internal/auth"
	mw "github.com/example/tasks-api/internal/middleware"
	"github.com/example/tasks-api/internal/postgres"
	"github.com/example/tasks-api/internal/server"
	"github.com/example/tasks-api/internal/task"
	"github.com/example/tasks-api/internal/user"
)

// TestAPIIntegration runs the whole stack (router → handlers → services →
// PostgreSQL) against a real database. It is skipped unless
// TEST_DATABASE_URL is set, e.g.:
//
//	TEST_DATABASE_URL=postgres://tasks:tasks@localhost:5433/tasks?sslmode=disable go test ./...
func TestAPIIntegration(t *testing.T) {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set")
	}
	ctx := t.Context()

	pool, err := postgres.Open(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}

	tokens, err := auth.NewTokenManager("integration-test-secret-32-bytes-long!", "tasks-api", "tasks-api", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	userSvc, err := user.NewService(postgres.NewUserRepo(pool), tokens)
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(server.NewHandler(server.Deps{
		Logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		DB:             pool,
		Tokens:         tokens,
		Users:          user.NewHandler(userSvc),
		Tasks:          task.NewHandler(task.NewService(postgres.NewTaskRepo(pool))),
		AuthLimiter:    mw.NewRateLimiter(100, time.Minute),
		RequestTimeout: 5 * time.Second,
	}))
	t.Cleanup(srv.Close)

	call := func(method, path, token string, body any) (int, map[string]any) {
		t.Helper()
		var rd io.Reader
		if body != nil {
			b, _ := json.Marshal(body)
			rd = bytes.NewReader(b)
		}
		req, _ := http.NewRequestWithContext(ctx, method, srv.URL+path, rd)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var out map[string]any
		_ = json.NewDecoder(res.Body).Decode(&out)
		return res.StatusCode, out
	}

	// Unique email per run so the test can be repeated on the same database.
	email := "it-" + time.Now().Format("20060102150405.000000000") + "@example.com"

	if code, body := call("POST", "/v1/auth/register", "", map[string]string{
		"email": email, "name": "Integration", "password": "password123",
	}); code != http.StatusCreated {
		t.Fatalf("register: %d %v", code, body)
	}
	if code, _ := call("POST", "/v1/auth/register", "", map[string]string{
		"email": email, "name": "Again", "password": "password123",
	}); code != http.StatusConflict {
		t.Fatalf("duplicate register: %d, want 409", code)
	}
	if code, _ := call("POST", "/v1/auth/login", "", map[string]string{
		"email": email, "password": "wrong-password",
	}); code != http.StatusUnauthorized {
		t.Fatalf("bad login: %d, want 401", code)
	}

	code, body := call("POST", "/v1/auth/login", "", map[string]string{"email": email, "password": "password123"})
	if code != http.StatusOK {
		t.Fatalf("login: %d %v", code, body)
	}
	token := body["data"].(map[string]any)["access_token"].(string)

	if code, _ := call("GET", "/v1/tasks", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("no token: %d, want 401", code)
	}

	for _, title := range []string{"one", "two", "three"} {
		if code, body := call("POST", "/v1/tasks", token, map[string]string{"title": title}); code != http.StatusCreated {
			t.Fatalf("create %s: %d %v", title, code, body)
		}
	}

	// Page through with limit=2: expect 2 + 1 items, newest first.
	code, body = call("GET", "/v1/tasks?limit=2", token, nil)
	if code != http.StatusOK {
		t.Fatalf("list: %d %v", code, body)
	}
	page1 := body["data"].([]any)
	pg := body["pagination"].(map[string]any)
	if len(page1) != 2 || pg["has_more"] != true {
		t.Fatalf("page 1 = %v", body)
	}
	if first := page1[0].(map[string]any)["title"]; first != "three" {
		t.Errorf("first item = %v, want three", first)
	}

	code, body = call("GET", "/v1/tasks?limit=2&cursor="+pg["next_cursor"].(string), token, nil)
	if code != http.StatusOK {
		t.Fatalf("list page 2: %d %v", code, body)
	}
	page2 := body["data"].([]any)
	if len(page2) != 1 || body["pagination"].(map[string]any)["has_more"] != false {
		t.Fatalf("page 2 = %v", body)
	}

	id := page2[0].(map[string]any)["id"].(string)
	if code, body := call("PATCH", "/v1/tasks/"+id, token, map[string]string{"status": "done"}); code != http.StatusOK ||
		body["data"].(map[string]any)["status"] != "done" {
		t.Fatalf("patch: %d %v", code, body)
	}
	if code, _ := call("DELETE", "/v1/tasks/"+id, token, nil); code != http.StatusNoContent {
		t.Fatalf("delete: %d", code)
	}
	if code, _ := call("GET", "/v1/tasks/"+id, token, nil); code != http.StatusNotFound {
		t.Fatalf("get deleted: %d, want 404", code)
	}
}
