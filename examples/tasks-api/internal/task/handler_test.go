package task_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/tasks-api/internal/auth"
	"github.com/example/tasks-api/internal/httpx"
	"github.com/example/tasks-api/internal/task"
)

// newTestServer builds a router with the task routes, pretending the
// request was already authenticated as userID (via the X-Test-User header).
func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	h := task.NewHandler(task.NewService(&memRepo{}))

	mux := http.NewServeMux()
	mux.Handle("GET /v1/tasks", httpx.HandlerFunc(h.List))
	mux.Handle("POST /v1/tasks", httpx.HandlerFunc(h.Create))
	mux.Handle("GET /v1/tasks/{id}", httpx.HandlerFunc(h.Get))
	mux.Handle("PATCH /v1/tasks/{id}", httpx.HandlerFunc(h.Update))
	mux.Handle("DELETE /v1/tasks/{id}", httpx.HandlerFunc(h.Delete))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := auth.WithUserID(r.Context(), r.Header.Get("X-Test-User"))
		mux.ServeHTTP(w, r.WithContext(ctx))
	})
}

func do(t *testing.T, h http.Handler, user, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Test-User", user)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

type taskBody struct {
	Data struct {
		ID     string `json:"id"`
		Title  string `json:"title"`
		Status string `json:"status"`
	} `json:"data"`
}

type errBody struct {
	Error struct {
		Code    string `json:"code"`
		Details []struct {
			Field string `json:"field"`
		} `json:"details"`
	} `json:"error"`
}

func decode[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("decode %q: %v", w.Body.String(), err)
	}
	return v
}

func TestCreateAndGet(t *testing.T) {
	srv := newTestServer(t)

	w := do(t, srv, "alice", "POST", "/v1/tasks", `{"title":"  เขียน docs  "}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d, body %s", w.Code, w.Body)
	}
	created := decode[taskBody](t, w)
	if created.Data.Title != "เขียน docs" {
		t.Errorf("title = %q, want trimmed", created.Data.Title)
	}
	if created.Data.Status != "todo" {
		t.Errorf("status = %q, want todo", created.Data.Status)
	}
	if loc := w.Header().Get("Location"); loc != "/v1/tasks/"+created.Data.ID {
		t.Errorf("Location = %q", loc)
	}

	// Owner can read it.
	if w := do(t, srv, "alice", "GET", "/v1/tasks/"+created.Data.ID, ""); w.Code != http.StatusOK {
		t.Errorf("owner get: status %d", w.Code)
	}
	// Another user gets 404, not 403: don't reveal that the task exists.
	if w := do(t, srv, "bob", "GET", "/v1/tasks/"+created.Data.ID, ""); w.Code != http.StatusNotFound {
		t.Errorf("other user get: status %d, want 404", w.Code)
	}
}

func TestCreateValidation(t *testing.T) {
	srv := newTestServer(t)

	tests := []struct {
		name      string
		body      string
		wantField string
	}{
		{"empty title", `{"title":"   "}`, "title"},
		{"bad status", `{"title":"x","status":"later"}`, "status"},
		{"title too long", `{"title":"` + strings.Repeat("ก", 201) + `"}`, "title"},
		{"unknown field", `{"title":"x","priority":1}`, "priority"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(t, srv, "alice", "POST", "/v1/tasks", tt.body)
			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status %d, want 422; body %s", w.Code, w.Body)
			}
			e := decode[errBody](t, w)
			if e.Error.Code != "VALIDATION_ERROR" {
				t.Errorf("code = %q", e.Error.Code)
			}
			if len(e.Error.Details) == 0 || e.Error.Details[0].Field != tt.wantField {
				t.Errorf("details = %+v, want field %q", e.Error.Details, tt.wantField)
			}
		})
	}
}

func TestUpdateAndDelete(t *testing.T) {
	srv := newTestServer(t)
	id := decode[taskBody](t, do(t, srv, "alice", "POST", "/v1/tasks", `{"title":"a"}`)).Data.ID

	w := do(t, srv, "alice", "PATCH", "/v1/tasks/"+id, `{"status":"done"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("patch: status %d, body %s", w.Code, w.Body)
	}
	got := decode[taskBody](t, w)
	if got.Data.Status != "done" || got.Data.Title != "a" {
		t.Errorf("after patch = %+v", got.Data)
	}

	if w := do(t, srv, "alice", "PATCH", "/v1/tasks/"+id, `{}`); w.Code != http.StatusUnprocessableEntity {
		t.Errorf("empty patch: status %d, want 422", w.Code)
	}

	if w := do(t, srv, "alice", "DELETE", "/v1/tasks/"+id, ""); w.Code != http.StatusNoContent {
		t.Errorf("delete: status %d, want 204", w.Code)
	}
	if w := do(t, srv, "alice", "DELETE", "/v1/tasks/"+id, ""); w.Code != http.StatusNotFound {
		t.Errorf("second delete: status %d, want 404", w.Code)
	}
	if w := do(t, srv, "alice", "GET", "/v1/tasks/not-a-uuid", ""); w.Code != http.StatusNotFound {
		t.Errorf("invalid id: status %d, want 404", w.Code)
	}
}

func TestListValidation(t *testing.T) {
	srv := newTestServer(t)
	for _, q := range []string{"limit=0", "limit=101", "limit=abc", "status=nope", "cursor=bm9wZQ"} {
		t.Run(q, func(t *testing.T) {
			if w := do(t, srv, "alice", "GET", "/v1/tasks?"+q, ""); w.Code != http.StatusUnprocessableEntity {
				t.Errorf("status %d, want 422", w.Code)
			}
		})
	}
}
