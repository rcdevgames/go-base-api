package http

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/usecase"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/httpresp"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/validation"
)

// Handler exposes auth HTTP endpoints.
type Handler struct {
	service usecase.Service
}

// NewHandler wires dependencies.
func NewHandler(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers auth routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
	})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login authenticates a user and issues tokens.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresp.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}
	validationResult := validation.ValidateStrings(map[string]string{
		"email":    req.Email,
		"password": req.Password,
	}, []validation.Rule{
		{Name: "email", Required: true, Pattern: regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)},
		{Name: "password", Required: true, MinLength: 6},
	})
	if !validationResult.IsValid() {
		httpresp.JSON(w, http.StatusBadRequest, validationResult.Errors)
		return
	}
	pair, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpresp.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	httpresp.JSON(w, http.StatusOK, pair)
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Refresh exchanges a refresh token for a new token pair.
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresp.Error(w, http.StatusBadRequest, "invalid payload")
		return
	}
	validationResult := validation.ValidateStrings(map[string]string{
		"refresh_token": req.RefreshToken,
	}, []validation.Rule{
		{Name: "refresh_token", Required: true, MinLength: 10},
	})
	if !validationResult.IsValid() {
		httpresp.JSON(w, http.StatusBadRequest, validationResult.Errors)
		return
	}
	pair, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpresp.Error(w, http.StatusUnauthorized, err.Error())
		return
	}
	httpresp.JSON(w, http.StatusOK, pair)
}
