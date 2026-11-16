// Package httpx contains the small HTTP helpers shared by every handler:
// the response envelope, error rendering, JSON decoding and validation.
package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Meta is attached to every successful response.
type Meta struct {
	RequestID string `json:"request_id,omitempty"`
}

// Pagination describes a cursor-paginated list.
type Pagination struct {
	Limit      int    `json:"limit"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

type envelope struct {
	Data       any         `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       Meta        `json:"meta"`
}

type errorBody struct {
	Code      string       `json:"code"`
	Message   string       `json:"message"`
	Details   []FieldError `json:"details,omitempty"`
	RequestID string       `json:"request_id,omitempty"`
}

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false) // keep "<" and "&" readable; this is an API, not HTML
	if err := enc.Encode(v); err != nil {
		// The status line is already sent; all we can do is log.
		slog.Error("encode response", "err", err)
	}
}

// OK writes {"data": ..., "meta": ...}.
func OK(w http.ResponseWriter, r *http.Request, status int, data any) {
	JSON(w, status, envelope{Data: data, Meta: Meta{RequestID: RequestID(r.Context())}})
}

// List writes {"data": [...], "pagination": ..., "meta": ...}.
func List(w http.ResponseWriter, r *http.Request, data any, p Pagination) {
	JSON(w, http.StatusOK, envelope{Data: data, Pagination: &p, Meta: Meta{RequestID: RequestID(r.Context())}})
}

// NoContent writes 204 with no body (used by DELETE).
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error renders err using the standard error envelope.
// Unknown errors become 500 and are logged; their text is never sent to the client.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	var apiErr *APIError
	switch {
	case errors.As(err, &apiErr):
		// Already describes itself.
	case errors.Is(err, context.DeadlineExceeded):
		Logger(r.Context()).Warn("request timed out", "err", err)
		apiErr = &APIError{
			Status:  http.StatusServiceUnavailable,
			Code:    CodeTimeout,
			Message: "request ใช้เวลานานเกินไป กรุณาลองใหม่",
		}
	default:
		Logger(r.Context()).Error("unhandled error", "err", err)
		apiErr = &APIError{
			Status:  http.StatusInternalServerError,
			Code:    CodeInternal,
			Message: "เกิดข้อผิดพลาดภายในระบบ",
		}
	}
	JSON(w, apiErr.Status, map[string]errorBody{
		"error": {
			Code:      apiErr.Code,
			Message:   apiErr.Message,
			Details:   apiErr.Details,
			RequestID: RequestID(r.Context()),
		},
	})
}

// HandlerFunc is an http.HandlerFunc that may return an error.
// Returning errors (instead of writing them in every handler) keeps
// error rendering in exactly one place.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		Error(w, r, err)
	}
}
