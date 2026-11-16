// Package middleware holds cross-cutting HTTP concerns.
// Every middleware has the standard shape func(http.Handler) http.Handler,
// so it works with any net/http compatible router.
package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"runtime/debug"
	"slices"
	"strings"
	"time"

	"github.com/example/tasks-api/internal/httpx"
)

// Middleware wraps an http.Handler with extra behaviour.
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares so that the first one listed is the outermost:
// Chain(h, A, B) == A(B(h)). A request flows A → B → h.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

// RequestID reuses a sane incoming X-Request-Id (e.g. from a load balancer)
// or generates a new one, stores it in the context and echoes it back.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-Id")
		if !validRequestID.MatchString(id) {
			id = newRequestID()
		}
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r.WithContext(httpx.WithRequestID(r.Context(), id)))
	})
}

func newRequestID() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b) // crypto/rand.Read never returns an error since Go 1.24
	return "req_" + hex.EncodeToString(b)
}

// statusRecorder captures the status code and size written by the handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (rec *statusRecorder) WriteHeader(code int) {
	if rec.status == 0 {
		rec.status = code
	}
	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}
	n, err := rec.ResponseWriter.Write(b)
	rec.bytes += n
	return n, err
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (rec *statusRecorder) Unwrap() http.ResponseWriter { return rec.ResponseWriter }

// Logger stores a request-scoped logger (tagged with the request ID) in the
// context and writes one structured access-log line per request.
// It must run after RequestID.
func Logger(base *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log := base.With("request_id", httpx.RequestID(r.Context()))
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r.WithContext(httpx.WithLogger(r.Context(), log)))

			if rec.status == 0 {
				rec.status = http.StatusOK
			}
			level := slog.LevelInfo
			if rec.status >= 500 {
				level = slog.LevelError
			}
			// Log the path only: query strings may contain sensitive data.
			log.LogAttrs(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rec.status),
				slog.Int("bytes", rec.bytes),
				slog.Duration("duration", time.Since(start)),
			)
		})
	}
}

// Recoverer turns a panic in a handler into a 500 response instead of
// letting it kill the connection, and logs the stack trace server-side.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			v := recover()
			if v == nil {
				return
			}
			if v == http.ErrAbortHandler { // deliberate abort; let net/http handle it
				panic(v)
			}
			httpx.Logger(r.Context()).Error("panic recovered",
				"panic", fmt.Sprint(v),
				"stack", string(debug.Stack()),
			)
			w.Header().Set("Connection", "close")
			httpx.Error(w, r, fmt.Errorf("panic: %v", v))
		}()
		next.ServeHTTP(w, r)
	})
}

// CORS allows browsers on the listed origins to call the API.
// Use exact origins in production; "*" is only for local development.
func CORS(allowedOrigins []string) Middleware {
	allowAll := slices.Contains(allowedOrigins, "*")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			w.Header().Add("Vary", "Origin")

			if origin != "" && (allowAll || slices.Contains(allowedOrigins, origin)) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Expose-Headers", "X-Request-Id, Location, Retry-After, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")

				// Preflight request.
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id")
					w.Header().Set("Access-Control-Max-Age", "600")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// SecureHeaders sets conservative security headers suitable for a JSON API.
func SecureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		next.ServeHTTP(w, r)
	})
}

// Timeout gives every request a deadline. Context-aware work (database calls,
// outgoing HTTP) stops when the deadline passes, and httpx.Error renders the
// resulting context.DeadlineExceeded as 503.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ContentTypeJSON rejects write requests whose body is not JSON (415).
func ContentTypeJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			ct := r.Header.Get("Content-Type")
			if !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				httpx.Error(w, r, &httpx.APIError{
					Status:  http.StatusUnsupportedMediaType,
					Code:    httpx.CodeUnsupportedMediaType,
					Message: "Content-Type ต้องเป็น application/json",
				})
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
