package user

import (
	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/registry"
	userhttp "github.com/rcdevgames/modular-monolith-clean/internal/modules/user/handler/http"
	userrepo "github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository/postgres"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/usecase"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"
)

// Module wires all components belonging to the user bounded context.
type Module struct {
	handler *userhttp.Handler
}

func init() {
	registry.Register("user", func(c *container.Container) (registry.Module, error) {
		return NewModule(c)
	})
}

// NewModule assembles the module dependencies.
func NewModule(c *container.Container) (*Module, error) {
	repo := userrepo.NewRepository(c.DB())
	service := usecase.NewService(repo)
	handler := userhttp.NewHandler(service)
	return &Module{handler: handler}, nil
}

// Name returns the module identifier.
func (m *Module) Name() string {
	return "user"
}

// RegisterRoutes exposes HTTP endpoints for the module.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
