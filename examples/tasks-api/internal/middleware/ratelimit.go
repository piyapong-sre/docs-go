package middleware

import (
	"context"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/example/tasks-api/internal/httpx"
)

// RateLimiter is a fixed-window limiter keyed by client IP.
//
// It keeps counters in process memory, which is fine for a single instance.
// When you run several replicas, move the counters to shared storage
// (e.g. Redis) or each replica will allow `limit` requests on its own.
type RateLimiter struct {
	limit  int
	window time.Duration
	now    func() time.Time

	mu      sync.Mutex
	clients map[string]*windowCounter
}

type windowCounter struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter allows limit requests per window for each client IP.
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		now:     time.Now,
		clients: make(map[string]*windowCounter),
	}
}

// allow records one request for key and reports whether it is permitted,
// how many requests remain, and when the current window resets.
func (rl *RateLimiter) allow(key string) (ok bool, remaining int, resetAt time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := rl.now()
	c, found := rl.clients[key]
	if !found || !now.Before(c.resetAt) {
		c = &windowCounter{resetAt: now.Add(rl.window)}
		rl.clients[key] = c
	}
	if c.count >= rl.limit {
		return false, 0, c.resetAt
	}
	c.count++
	return true, rl.limit - c.count, c.resetAt
}

// Cleanup removes expired windows every interval until ctx is cancelled,
// so the map does not grow forever.
func (rl *RateLimiter) Cleanup(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			rl.mu.Lock()
			now := rl.now()
			for k, c := range rl.clients {
				if !now.Before(c.resetAt) {
					delete(rl.clients, k)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// Middleware enforces the limit and reports quota in response headers.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ok, remaining, resetAt := rl.allow(clientIP(r))

		h := w.Header()
		h.Set("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		h.Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
		h.Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

		if !ok {
			retryAfter := max(1, int(resetAt.Sub(rl.now()).Seconds()+0.5))
			h.Set("Retry-After", strconv.Itoa(retryAfter))
			httpx.Error(w, r, &httpx.APIError{
				Status:  http.StatusTooManyRequests,
				Code:    httpx.CodeRateLimitExceeded,
				Message: "เรียกใช้บ่อยเกินไป กรุณารอ " + strconv.Itoa(retryAfter) + " วินาที",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// clientIP uses the TCP peer address. Do NOT trust X-Forwarded-For unless
// the API sits behind a proxy you control that overwrites that header.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
