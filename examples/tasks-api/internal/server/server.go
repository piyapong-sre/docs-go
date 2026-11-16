// Package server wires handlers and middleware into an http.Handler and
// configures the *http.Server.
package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/example/tasks-api/api"
	"github.com/example/tasks-api/internal/auth"
	"github.com/example/tasks-api/internal/httpx"
	mw "github.com/example/tasks-api/internal/middleware"
	"github.com/example/tasks-api/internal/task"
	"github.com/example/tasks-api/internal/user"
)

// Pinger reports whether a dependency (the database) is reachable.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Deps is everything the routes need. Passing dependencies explicitly
// (no globals) makes the wiring obvious and the server easy to test.
type Deps struct {
	Logger         *slog.Logger
	DB             Pinger
	Tokens         auth.Verifier
	Users          *user.Handler
	Tasks          *task.Handler
	AuthLimiter    *mw.RateLimiter
	CORSOrigins    []string
	RequestTimeout time.Duration
}

// NewHandler returns the root handler with every route registered.
func NewHandler(d Deps) http.Handler {
	mux := http.NewServeMux()
	requireAuth := auth.RequireAuth(d.Tokens)

	// Operational endpoints (not versioned).
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := d.DB.Ping(ctx); err != nil {
			httpx.Logger(r.Context()).Warn("readiness check failed", "err", err)
			httpx.JSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
		httpx.JSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(api.OpenAPISpec)
	})

	// Auth: stricter rate limit against brute-force attempts.
	limited := d.AuthLimiter.Middleware
	mux.Handle("POST /v1/auth/register", limited(httpx.HandlerFunc(d.Users.Register)))
	mux.Handle("POST /v1/auth/login", limited(httpx.HandlerFunc(d.Users.Login)))
	mux.Handle("GET /v1/me", requireAuth(httpx.HandlerFunc(d.Users.Me)))

	// Tasks: every route requires a logged-in user.
	mux.Handle("GET /v1/tasks", requireAuth(httpx.HandlerFunc(d.Tasks.List)))
	mux.Handle("POST /v1/tasks", requireAuth(httpx.HandlerFunc(d.Tasks.Create)))
	mux.Handle("GET /v1/tasks/{id}", requireAuth(httpx.HandlerFunc(d.Tasks.Get)))
	mux.Handle("PATCH /v1/tasks/{id}", requireAuth(httpx.HandlerFunc(d.Tasks.Update)))
	mux.Handle("DELETE /v1/tasks/{id}", requireAuth(httpx.HandlerFunc(d.Tasks.Delete)))

	// Anything else: JSON 404 instead of the default plain-text page.
	mux.Handle("/", httpx.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return httpx.NotFound("ไม่พบ endpoint นี้")
	}))

	// Global middleware, outermost first.
	return mw.Chain(mux,
		mw.RequestID,
		mw.Logger(d.Logger),
		mw.Recoverer,
		mw.SecureHeaders,
		mw.CORS(d.CORSOrigins),
		mw.Timeout(d.RequestTimeout),
		mw.ContentTypeJSON,
	)
}

// New returns an *http.Server with production-safe timeouts.
// The zero-value http.Server has NO timeouts, which lets slow clients
// hold connections open forever.
func New(addr string, h http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
	}
}
