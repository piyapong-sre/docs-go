package httpx

import (
	"fmt"
	"net/http"
)

// Error codes are stable strings that clients can switch on.
const (
	CodeBadRequest           = "BAD_REQUEST"
	CodeUnauthenticated      = "UNAUTHENTICATED"
	CodeForbidden            = "FORBIDDEN"
	CodeNotFound             = "NOT_FOUND"
	CodeMethodNotAllowed     = "METHOD_NOT_ALLOWED"
	CodeConflict             = "CONFLICT"
	CodePayloadTooLarge      = "PAYLOAD_TOO_LARGE"
	CodeUnsupportedMediaType = "UNSUPPORTED_MEDIA_TYPE"
	CodeValidation           = "VALIDATION_ERROR"
	CodeRateLimitExceeded    = "RATE_LIMIT_EXCEEDED"
	CodeTimeout              = "TIMEOUT"
	CodeInternal             = "INTERNAL_ERROR"
)

// FieldError describes why a single input field was rejected.
type FieldError struct {
	Field string `json:"field"`
	Issue string `json:"issue"`
}

// APIError is an error that knows how it should be rendered over HTTP.
// Handlers return it (or a plain error, which becomes a 500).
type APIError struct {
	Status  int
	Code    string
	Message string
	Details []FieldError
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%d %s: %s", e.Status, e.Code, e.Message)
}

func BadRequest(msg string) *APIError {
	return &APIError{Status: http.StatusBadRequest, Code: CodeBadRequest, Message: msg}
}

func Unauthenticated(msg string) *APIError {
	return &APIError{Status: http.StatusUnauthorized, Code: CodeUnauthenticated, Message: msg}
}

func NotFound(msg string) *APIError {
	return &APIError{Status: http.StatusNotFound, Code: CodeNotFound, Message: msg}
}

func Conflict(msg string) *APIError {
	return &APIError{Status: http.StatusConflict, Code: CodeConflict, Message: msg}
}

func Validation(details []FieldError) *APIError {
	return &APIError{
		Status:  http.StatusUnprocessableEntity,
		Code:    CodeValidation,
		Message: "ข้อมูลไม่ผ่านการตรวจสอบ",
		Details: details,
	}
}
