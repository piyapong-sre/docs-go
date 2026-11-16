package httpx

import (
	"context"
	"log/slog"
)

// Unexported key types prevent collisions with other packages' context keys.
type (
	requestIDKey struct{}
	loggerKey    struct{}
)

// WithRequestID returns a copy of ctx carrying the request ID.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID returns the request ID stored in ctx, or "".
func RequestID(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey{}).(string)
	return id
}

// WithLogger returns a copy of ctx carrying a request-scoped logger.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

// Logger returns the request-scoped logger, falling back to slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}
