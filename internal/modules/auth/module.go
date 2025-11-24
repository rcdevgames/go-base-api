package auth

import (
	"github.com/go-chi/chi/v5"

	authhttp "github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/handler/http"
	authrepo "github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/repository/postgres"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/usecase"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/registry"
	userrepo "github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository/postgres"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"
)

type Module struct {
	handler *authhttp.Handler
}

func init() {
	registry.Register("auth", func(c *container.Container) (registry.Module, error) {
		return NewModule(c)
	})
}

// NewModule constructs the auth module.
func NewModule(c *container.Container) (registry.Module, error) {
	userRepository := userrepo.NewRepository(c.DB())
	authRepository := authrepo.NewRepository(userRepository)
	securityCfg := c.Config().Security
	service := usecase.NewService(
		authRepository,
		securityCfg.JWTSecret,
		securityCfg.JWTExpiration,
		securityCfg.RefreshSecret,
		securityCfg.RefreshExpiration,
	)
	handler := authhttp.NewHandler(service)
	return &Module{handler: handler}, nil
}

// Name returns module name.
func (m *Module) Name() string { return "auth" }

// RegisterRoutes mounts auth routes.
func (m *Module) RegisterRoutes(r chi.Router) {
	m.handler.RegisterRoutes(r)
}
