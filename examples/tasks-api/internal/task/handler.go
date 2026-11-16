package task

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/example/tasks-api/internal/auth"
	"github.com/example/tasks-api/internal/httpx"
)

const (
	maxTitleLen       = 200
	maxDescriptionLen = 5000
)

// Handler is the HTTP transport for tasks. It translates between JSON and
// the domain, validates input and maps domain errors to HTTP errors.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ---- Response DTO ----------------------------------------------------------

// taskResponse is the public JSON shape of a task. Keeping it separate from
// the domain model means internal fields can never leak by accident.
type taskResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	DueAt       *time.Time `json:"due_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toResponse(t Task) taskResponse {
	r := taskResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		CreatedAt:   t.CreatedAt.UTC(),
		UpdatedAt:   t.UpdatedAt.UTC(),
	}
	if t.DueAt != nil {
		due := t.DueAt.UTC()
		r.DueAt = &due
	}
	return r
}

// ---- Create ----------------------------------------------------------------

type createRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      Status     `json:"status"`
	DueAt       *time.Time `json:"due_at"`
}

func (req *createRequest) validate() error {
	req.Title = strings.TrimSpace(req.Title)

	var v httpx.Validator
	v.Check(req.Title != "", "title", "ห้ามว่าง")
	v.Check(utf8.RuneCountInString(req.Title) <= maxTitleLen, "title", "ยาวได้ไม่เกิน 200 ตัวอักษร")
	v.Check(utf8.RuneCountInString(req.Description) <= maxDescriptionLen, "description", "ยาวได้ไม่เกิน 5000 ตัวอักษร")
	v.Check(req.Status == "" || req.Status.Valid(), "status", "ต้องเป็น todo, in_progress หรือ done")
	return v.Err()
}

// Create handles POST /v1/tasks.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) error {
	userID := mustUserID(r)

	var req createRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	if err := req.validate(); err != nil {
		return err
	}

	t, err := h.svc.Create(r.Context(), CreateParams{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueAt:       req.DueAt,
	})
	if err != nil {
		return err
	}

	w.Header().Set("Location", "/v1/tasks/"+t.ID)
	httpx.OK(w, r, http.StatusCreated, toResponse(t))
	return nil
}

// ---- Get -------------------------------------------------------------------

// Get handles GET /v1/tasks/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) error {
	id, ok := httpx.PathUUID(r, "id")
	if !ok {
		return errTaskNotFound
	}
	t, err := h.svc.Get(r.Context(), mustUserID(r), id)
	if err != nil {
		return mapError(err)
	}
	httpx.OK(w, r, http.StatusOK, toResponse(t))
	return nil
}

// ---- List ------------------------------------------------------------------

// List handles GET /v1/tasks?status=todo&limit=20&cursor=...
func (h *Handler) List(w http.ResponseWriter, r *http.Request) error {
	q := r.URL.Query()
	p := ListParams{UserID: mustUserID(r)}

	var v httpx.Validator
	if s := q.Get("status"); s != "" {
		st := Status(s)
		v.Check(st.Valid(), "status", "ต้องเป็น todo, in_progress หรือ done")
		p.Status = &st
	}
	if s := q.Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		v.Check(err == nil && n >= 1 && n <= MaxPageSize, "limit", "ต้องเป็นตัวเลข 1-100")
		p.Limit = n
	}
	if s := q.Get("cursor"); s != "" {
		c, err := DecodeCursor(s)
		v.Check(err == nil && httpx.IsUUID(c.ID), "cursor", "cursor ไม่ถูกต้อง")
		p.After = &c
	}
	if err := v.Err(); err != nil {
		return err
	}

	page, err := h.svc.List(r.Context(), p)
	if err != nil {
		return err
	}

	items := make([]taskResponse, 0, len(page.Tasks))
	for _, t := range page.Tasks {
		items = append(items, toResponse(t))
	}
	pg := httpx.Pagination{Limit: clampLimit(p.Limit), HasMore: page.Next != nil}
	if page.Next != nil {
		pg.NextCursor = EncodeCursor(*page.Next)
	}
	httpx.List(w, r, items, pg)
	return nil
}

// ---- Update ----------------------------------------------------------------

// updateRequest uses pointers so we can tell "field omitted" (nil)
// from "field set to the zero value" ("" or false).
type updateRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *Status    `json:"status"`
	DueAt       *time.Time `json:"due_at"`
}

func (req *updateRequest) validate() error {
	var v httpx.Validator
	v.Check(req.Title != nil || req.Description != nil || req.Status != nil || req.DueAt != nil,
		"body", "ต้องส่งอย่างน้อย 1 field")
	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		req.Title = &t
		v.Check(t != "", "title", "ห้ามว่าง")
		v.Check(utf8.RuneCountInString(t) <= maxTitleLen, "title", "ยาวได้ไม่เกิน 200 ตัวอักษร")
	}
	if req.Description != nil {
		v.Check(utf8.RuneCountInString(*req.Description) <= maxDescriptionLen, "description", "ยาวได้ไม่เกิน 5000 ตัวอักษร")
	}
	if req.Status != nil {
		v.Check(req.Status.Valid(), "status", "ต้องเป็น todo, in_progress หรือ done")
	}
	return v.Err()
}

// Update handles PATCH /v1/tasks/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) error {
	id, ok := httpx.PathUUID(r, "id")
	if !ok {
		return errTaskNotFound
	}
	var req updateRequest
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		return err
	}
	if err := req.validate(); err != nil {
		return err
	}

	t, err := h.svc.Update(r.Context(), UpdateParams{
		ID:          id,
		UserID:      mustUserID(r),
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		DueAt:       req.DueAt,
	})
	if err != nil {
		return mapError(err)
	}
	httpx.OK(w, r, http.StatusOK, toResponse(t))
	return nil
}

// ---- Delete ----------------------------------------------------------------

// Delete handles DELETE /v1/tasks/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) error {
	id, ok := httpx.PathUUID(r, "id")
	if !ok {
		return errTaskNotFound
	}
	if err := h.svc.Delete(r.Context(), mustUserID(r), id); err != nil {
		return mapError(err)
	}
	httpx.NoContent(w)
	return nil
}

// ---- helpers ---------------------------------------------------------------

var errTaskNotFound = httpx.NotFound("ไม่พบ task")

// mapError converts domain errors to HTTP errors. Anything unknown is
// returned unchanged and becomes a 500.
func mapError(err error) error {
	if errors.Is(err, ErrNotFound) {
		return errTaskNotFound
	}
	return err
}

// mustUserID reads the user set by auth.RequireAuth. Routes using this
// handler are always wrapped by that middleware, so a missing ID is a
// programming error.
func mustUserID(r *http.Request) string {
	id, ok := auth.UserID(r.Context())
	if !ok {
		panic("task handler used without auth.RequireAuth")
	}
	return id
}
