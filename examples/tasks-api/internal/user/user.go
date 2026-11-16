// Package user handles registration, login and the current user's profile.
package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/example/tasks-api/internal/auth"
)

type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	CreatedAt    time.Time
}

var (
	ErrNotFound           = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Repository is the storage contract for users.
type Repository interface {
	Create(ctx context.Context, email, name, passwordHash string) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id string) (User, error)
}

// TokenIssuer creates access tokens. *auth.TokenManager implements it.
type TokenIssuer interface {
	Issue(userID string) (token string, expiresAt time.Time, err error)
}

type Service struct {
	repo   Repository
	tokens TokenIssuer

	// dummyHash is compared against when the email does not exist, so a
	// login for an unknown email takes as long as one with a wrong password.
	// This prevents attackers from discovering registered emails by timing.
	dummyHash string
}

func NewService(repo Repository, tokens TokenIssuer) (*Service, error) {
	h, err := auth.HashPassword("dummy-password-for-timing")
	if err != nil {
		return nil, err
	}
	return &Service{repo: repo, tokens: tokens, dummyHash: h}, nil
}

// NormalizeEmail trims spaces and lower-cases an email address.
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (s *Service) Register(ctx context.Context, email, name, password string) (User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return User{}, fmt.Errorf("hash password: %w", err)
	}
	u, err := s.repo.Create(ctx, NormalizeEmail(email), strings.TrimSpace(name), hash)
	if err != nil {
		return User{}, err // ErrEmailTaken passes through unchanged
	}
	return u, nil
}

// Token is the result of a successful login.
type Token struct {
	AccessToken string
	ExpiresAt   time.Time
}

func (s *Service) Login(ctx context.Context, email, password string) (Token, error) {
	u, err := s.repo.GetByEmail(ctx, NormalizeEmail(email))
	switch {
	case errors.Is(err, ErrNotFound):
		_, _ = auth.CheckPassword(s.dummyHash, password) // equalize timing
		return Token{}, ErrInvalidCredentials
	case err != nil:
		return Token{}, fmt.Errorf("get user: %w", err)
	}

	ok, err := auth.CheckPassword(u.PasswordHash, password)
	if err != nil {
		return Token{}, fmt.Errorf("check password: %w", err)
	}
	if !ok {
		return Token{}, ErrInvalidCredentials
	}

	tok, exp, err := s.tokens.Issue(u.ID)
	if err != nil {
		return Token{}, err
	}
	return Token{AccessToken: tok, ExpiresAt: exp}, nil
}

func (s *Service) Get(ctx context.Context, id string) (User, error) {
	return s.repo.GetByID(ctx, id)
}
