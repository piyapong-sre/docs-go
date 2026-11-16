package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-that-is-at-least-32-bytes-long"

func TestTokenRoundTrip(t *testing.T) {
	tm, err := NewTokenManager(testSecret, "tasks-api", "tasks-api", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	tok, _, err := tm.Issue("user-123")
	if err != nil {
		t.Fatal(err)
	}
	got, err := tm.Verify(tok)
	if err != nil {
		t.Fatal(err)
	}
	if got != "user-123" {
		t.Errorf("subject = %q, want user-123", got)
	}
}

func TestTokenRejected(t *testing.T) {
	now := time.Now()
	tm, _ := NewTokenManager(testSecret, "tasks-api", "tasks-api", time.Minute)
	tm.now = func() time.Time { return now }
	tok, _, _ := tm.Issue("user-123")

	other, _ := NewTokenManager("another-secret-that-is-also-32-bytes!!", "tasks-api", "tasks-api", time.Minute)
	otherIssuer, _ := NewTokenManager(testSecret, "someone-else", "tasks-api", time.Minute)

	expired, _ := NewTokenManager(testSecret, "tasks-api", "tasks-api", time.Minute)
	expired.now = func() time.Time { return now.Add(2 * time.Minute) }

	tests := map[string]struct {
		tm    *TokenManager
		token string
	}{
		"wrong secret":  {other, tok},
		"wrong issuer":  {otherIssuer, tok},
		"expired":       {expired, tok},
		"garbage":       {tm, "not.a.jwt"},
		"alg none":      {tm, "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzdWIiOiJ1c2VyLTEyMyJ9."},
		"tampered body": {tm, tok[:len(tok)-2] + "xx"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := tt.tm.Verify(tt.token); err == nil {
				t.Error("expected token to be rejected")
			}
		})
	}
}

func TestNewTokenManagerShortSecret(t *testing.T) {
	if _, err := NewTokenManager("short", "i", "a", time.Minute); err == nil {
		t.Error("expected error for short secret")
	}
}
