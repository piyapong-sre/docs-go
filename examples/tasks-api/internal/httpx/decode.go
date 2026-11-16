package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// MaxBodyBytes limits the size of JSON request bodies (1 MB).
const MaxBodyBytes = 1 << 20

// DecodeJSON reads a single JSON object from the request body into dst.
// It rejects unknown fields, trailing data and bodies larger than MaxBodyBytes,
// and turns every failure into a client-friendly *APIError.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var (
			syntaxErr   *json.SyntaxError
			typeErr     *json.UnmarshalTypeError
			maxBytesErr *http.MaxBytesError
		)
		switch {
		case errors.As(err, &syntaxErr):
			return BadRequest(fmt.Sprintf("JSON ไม่ถูกต้องที่ตำแหน่ง %d", syntaxErr.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return BadRequest("JSON ไม่สมบูรณ์")
		case errors.As(err, &typeErr):
			return Validation([]FieldError{{
				Field: typeErr.Field,
				Issue: fmt.Sprintf("ต้องเป็นชนิด %s", typeErr.Type),
			}})
		case errors.Is(err, io.EOF):
			return BadRequest("request body ว่าง")
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			field := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
			return Validation([]FieldError{{Field: field, Issue: "ไม่รู้จัก field นี้"}})
		case errors.As(err, &maxBytesErr):
			return &APIError{
				Status:  http.StatusRequestEntityTooLarge,
				Code:    CodePayloadTooLarge,
				Message: fmt.Sprintf("request body ต้องไม่เกิน %d bytes", maxBytesErr.Limit),
			}
		default:
			return err
		}
	}

	// A body like `{"a":1}{"b":2}` contains more than one value.
	if dec.More() {
		return BadRequest("request body ต้องมี JSON object เดียว")
	}
	return nil
}

var uuidRE = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// IsUUID reports whether s looks like a UUID (any version).
func IsUUID(s string) bool { return uuidRE.MatchString(s) }

// PathUUID reads a path parameter that must be a UUID.
// An invalid UUID can never match a row, so callers should answer 404.
func PathUUID(r *http.Request, name string) (string, bool) {
	v := r.PathValue(name)
	return strings.ToLower(v), IsUUID(v)
}
