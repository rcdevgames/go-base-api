package http

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/entity"
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

// Login authenticates a user and issues tokens.
// @Summary Login
// @Description Authenticate a user using email and password to obtain a token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body entity.LoginRequest true "Credentials"
// @Success 200 {object} httpresp.SuccessResponse{data=entity.TokenPair}
// @Failure 400 {object} httpresp.ErrorResponse{errors=validation.ErrorMap}
// @Failure 401 {object} httpresp.ErrorResponse
// @Failure 500 {object} httpresp.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req entity.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresp.Error(w, http.StatusBadRequest, "Invalid payload", nil)
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
		httpresp.Error(w, http.StatusBadRequest, "Validation error", validationResult.Errors)
		return
	}
	pair, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		httpresp.Error(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	httpresp.JSON(w, http.StatusOK, "Success", pair)
}

// Refresh exchanges a refresh token for a new token pair.
// @Summary Refresh Token
// @Description Exchange a valid refresh token for a new token pair.
// @Tags Auth
// @Accept json
// @Produce json
// @Param payload body entity.RefreshRequest true "Refresh token"
// @Success 200 {object} httpresp.SuccessResponse{data=entity.TokenPair}
// @Failure 400 {object} httpresp.ErrorResponse{errors=validation.ErrorMap}
// @Failure 401 {object} httpresp.ErrorResponse
// @Failure 500 {object} httpresp.ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req entity.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpresp.Error(w, http.StatusBadRequest, "Invalid payload", nil)
		return
	}
	validationResult := validation.ValidateStrings(map[string]string{
		"refresh_token": req.RefreshToken,
	}, []validation.Rule{
		{Name: "refresh_token", Required: true, MinLength: 10},
	})
	if !validationResult.IsValid() {
		httpresp.Error(w, http.StatusBadRequest, "Validation error", validationResult.Errors)
		return
	}
	pair, err := h.service.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		httpresp.Error(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	httpresp.JSON(w, http.StatusOK, "Success", pair)
}
