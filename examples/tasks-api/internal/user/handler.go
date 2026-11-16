package user

import (
	"errors"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/example/tasks-api/internal/auth"
	"github.com/example/tasks-api/internal/httpx"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// userResponse never includes the password hash.
type userResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

func toResponse(u User) userResponse {
	return userResponse{ID: u.ID, Email: u.Email, Name: u.Name, CreatedAt: u.CreatedAt.UTC()}
}

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (req *registerRequest) validate() error {
	req.Name = strings.TrimSpace(req.Name)

	var v httpx.Validator
	// ParseAddress also accepts "Name <a@b.com>", so require a bare address.
	addr, err := mail.ParseAddress(req.Email)
	v.Check(err == nil && addr.Address == req.Email && len(req.Email) <= 254, "email", "รูปแบบอีเมลไม่ถูกต้อง")
	v.Check(req.Name != "", "name", "ห้ามว่าง")
	v.Check(utf8.RuneCountInString(req.Name) <= 100, "name", "ยาวได้ไม่เกิน 100 ตัวอักษร")
	v.Check(utf8.RuneCountInString(req.Password) >= 8, "password", "ต้องยาวอย่างน้อย 8 ตัวอักษร")
	v.Check(len(req.Password) <= auth.MaxPasswordBytes, "password", "ยาวได้ไม่เกิน 72 bytes")
	return v.Err()
}

// Register handles POST /v1/auth/register.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) error {
	var req registerRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	if err := req.validate(); err != nil {
		return err
	}

	u, err := h.svc.Register(r.Context(), req.Email, req.Name, req.Password)
	if errors.Is(err, ErrEmailTaken) {
		return httpx.Conflict("อีเมลนี้ถูกใช้แล้ว")
	}
	if err != nil {
		return err
	}

	w.Header().Set("Location", "/v1/me")
	httpx.OK(w, r, http.StatusCreated, toResponse(u))
	return nil
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type tokenResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresIn   int       `json:"expires_in"` // seconds
	ExpiresAt   time.Time `json:"expires_at"`
}

// Login handles POST /v1/auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) error {
	var req loginRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	var v httpx.Validator
	v.Check(req.Email != "", "email", "ห้ามว่าง")
	v.Check(req.Password != "", "password", "ห้ามว่าง")
	if err := v.Err(); err != nil {
		return err
	}

	tok, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, ErrInvalidCredentials) {
		// Same message for "no such email" and "wrong password".
		return httpx.Unauthenticated("อีเมลหรือรหัสผ่านไม่ถูกต้อง")
	}
	if err != nil {
		return err
	}

	// Tokens must never be cached by browsers or proxies.
	w.Header().Set("Cache-Control", "no-store")
	httpx.OK(w, r, http.StatusOK, tokenResponse{
		AccessToken: tok.AccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(time.Until(tok.ExpiresAt).Round(time.Second).Seconds()),
		ExpiresAt:   tok.ExpiresAt.UTC(),
	})
	return nil
}

// Me handles GET /v1/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) error {
	id, ok := auth.UserID(r.Context())
	if !ok {
		return httpx.Unauthenticated("ต้องเข้าสู่ระบบ")
	}
	u, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		// The token is valid but the user was deleted.
		return httpx.Unauthenticated("ไม่พบผู้ใช้")
	}
	if err != nil {
		return err
	}
	httpx.OK(w, r, http.StatusOK, toResponse(u))
	return nil
}
