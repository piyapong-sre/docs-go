package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rl := NewRateLimiter(2, time.Minute)
	rl.now = func() time.Time { return now } // fake clock: no sleeping in tests

	h := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	do := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
		r.RemoteAddr = "203.0.113.7:51000"
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}

	for i := range 2 {
		if w := do(); w.Code != http.StatusOK {
			t.Fatalf("request %d: status %d, want 200", i+1, w.Code)
		}
	}

	w := do()
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("3rd request: status %d, want 429", w.Code)
	}
	if got := w.Header().Get("Retry-After"); got != "60" {
		t.Errorf("Retry-After = %q, want 60", got)
	}
	if got := w.Header().Get("X-RateLimit-Remaining"); got != "0" {
		t.Errorf("X-RateLimit-Remaining = %q, want 0", got)
	}

	// After the window resets the client may call again.
	now = now.Add(time.Minute)
	if w := do(); w.Code != http.StatusOK {
		t.Fatalf("after reset: status %d, want 200", w.Code)
	}
}
