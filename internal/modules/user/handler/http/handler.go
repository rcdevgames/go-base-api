package http

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/usecase"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/httpresp"
)

// Handler wraps HTTP specific logic for the user module.
type Handler struct {
	service usecase.Service
}

// NewHandler constructs a Handler instance.
func NewHandler(service usecase.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes mounts user endpoints under the provided router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", h.GetByID)
	})
}

// GetByID handles GET /users/{id} requests.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		httpresp.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}

	user, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			httpresp.Error(w, http.StatusNotFound, err.Error())
			return
		}
		httpresp.Error(w, http.StatusInternalServerError, "unable to fetch user")
		return
	}

	httpresp.JSON(w, http.StatusOK, user)
}
