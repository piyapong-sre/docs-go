package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/example/tasks-api/internal/httpx"
)

type userIDKey struct{}

// Verifier is the part of TokenManager the middleware needs.
// Depending on a small interface makes the middleware easy to test.
type Verifier interface {
	Verify(token string) (userID string, err error)
}

// RequireAuth rejects requests without a valid "Authorization: Bearer <token>"
// header and stores the authenticated user ID in the request context.
func RequireAuth(v Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
			if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
				w.Header().Set("WWW-Authenticate", `Bearer`)
				httpx.Error(w, r, httpx.Unauthenticated("ต้องส่ง Authorization: Bearer <token>"))
				return
			}
			userID, err := v.Verify(token)
			if err != nil {
				w.Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
				httpx.Error(w, r, httpx.Unauthenticated("token ไม่ถูกต้องหรือหมดอายุ"))
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserID returns the authenticated user's ID. It is only set on routes
// wrapped by RequireAuth.
func UserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey{}).(string)
	return id, ok && id != ""
}

// WithUserID is used by tests to simulate an authenticated request.
func WithUserID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, userIDKey{}, id)
}
