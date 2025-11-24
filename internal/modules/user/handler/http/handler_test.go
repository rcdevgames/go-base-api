package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository"
)

type stubService struct {
	user *entity.User
	err  error
}

func (s stubService) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestHandler_GetByID(t *testing.T) {
	h := NewHandler(stubService{user: &entity.User{ID: 1, Name: "Tester", Email: "tester@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}})
	req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandler_GetByID_InvalidID(t *testing.T) {
	h := NewHandler(stubService{})
	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "abc")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	h := NewHandler(stubService{err: repository.ErrNotFound})
	req := httptest.NewRequest(http.MethodGet, "/users/2", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "2")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_GetByID_InternalError(t *testing.T) {
	h := NewHandler(stubService{err: errors.New("boom")})
	req := httptest.NewRequest(http.MethodGet, "/users/3", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "3")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	w := httptest.NewRecorder()

	h.GetByID(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
