package httpx_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/example/tasks-api/internal/httpx"
)

func TestDecodeJSON(t *testing.T) {
	type input struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int // 0 means success
		wantField  string
	}{
		{name: "valid", body: `{"name":"สมชาย","age":30}`},
		{name: "empty body", body: ``, wantStatus: http.StatusBadRequest},
		{name: "syntax error", body: `{"name":}`, wantStatus: http.StatusBadRequest},
		{name: "truncated", body: `{"name":"a"`, wantStatus: http.StatusBadRequest},
		{name: "wrong type", body: `{"age":"thirty"}`, wantStatus: http.StatusUnprocessableEntity, wantField: "age"},
		{name: "unknown field", body: `{"nickname":"x"}`, wantStatus: http.StatusUnprocessableEntity, wantField: "nickname"},
		{name: "two objects", body: `{"name":"a"}{"name":"b"}`, wantStatus: http.StatusBadRequest},
		{name: "too large", body: `{"name":"` + strings.Repeat("a", httpx.MaxBodyBytes) + `"}`, wantStatus: http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			var dst input
			err := httpx.DecodeJSON(w, r, &dst)

			if tt.wantStatus == 0 {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			var apiErr *httpx.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *APIError, got %T: %v", err, err)
			}
			if apiErr.Status != tt.wantStatus {
				t.Errorf("status = %d, want %d (%s)", apiErr.Status, tt.wantStatus, apiErr.Message)
			}
			if tt.wantField != "" && (len(apiErr.Details) == 0 || apiErr.Details[0].Field != tt.wantField) {
				t.Errorf("details = %+v, want field %q", apiErr.Details, tt.wantField)
			}
		})
	}
}
