// Package auth issues and verifies access tokens and hashes passwords.
package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken is returned for any token that must not be trusted:
// bad signature, expired, wrong issuer/audience, malformed, ...
var ErrInvalidToken = errors.New("invalid token")

// TokenManager issues and verifies HS256-signed JWT access tokens.
type TokenManager struct {
	secret   []byte
	issuer   string
	audience string
	ttl      time.Duration
	now      func() time.Time
}

// NewTokenManager requires a secret of at least 32 bytes.
func NewTokenManager(secret, issuer, audience string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("jwt secret must be at least 32 bytes")
	}
	return &TokenManager{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		ttl:      ttl,
		now:      time.Now,
	}, nil
}

// Issue creates a signed token whose subject is userID.
func (tm *TokenManager) Issue(userID string) (token string, expiresAt time.Time, err error) {
	// JWT times have whole-second precision; truncate so expiresAt matches "exp".
	now := tm.now().Truncate(time.Second)
	expiresAt = now.Add(tm.ttl)
	claims := jwt.RegisteredClaims{
		Subject:   userID,
		Issuer:    tm.issuer,
		Audience:  jwt.ClaimStrings{tm.audience},
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	}
	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(tm.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}
	return token, expiresAt, nil
}

// Verify checks the token and returns the user ID (subject).
func (tm *TokenManager) Verify(token string) (string, error) {
	claims := &jwt.RegisteredClaims{}
	_, err := jwt.ParseWithClaims(token, claims,
		func(*jwt.Token) (any, error) { return tm.secret, nil },
		// Pin the algorithm: never let the token choose how it is verified.
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(tm.issuer),
		jwt.WithAudience(tm.audience),
		jwt.WithExpirationRequired(),
		jwt.WithTimeFunc(tm.now),
	)
	if err != nil || claims.Subject == "" {
		return "", ErrInvalidToken
	}
	return claims.Subject, nil
}
