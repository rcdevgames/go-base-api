package healthcheck

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Handler exposes liveness and readiness endpoints for the service.
type Handler struct {
	db *sql.DB
}

// NewHandler wires dependencies required to serve health information.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{db: db}
}

// RegisterRoutes mounts health endpoints on the given router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/healthz", h.Liveness)
	r.Get("/readyz", h.Readiness)
}

// Liveness indicates whether the process is responsive.
func (h *Handler) Liveness(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("OK"))
}

// Readiness verifies application dependencies (database connection).
func (h *Handler) Readiness(w http.ResponseWriter, _ *http.Request) {
	if err := h.db.Ping(); err != nil {
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("READY"))
}
