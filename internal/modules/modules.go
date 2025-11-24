package modules

import (
	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/registry"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"

	// Module registrations
	_ "github.com/rcdevgames/modular-monolith-clean/internal/modules/auth"
	_ "github.com/rcdevgames/modular-monolith-clean/internal/modules/user"
)

// RegisterAll discovers available modules and mounts their routes.
func RegisterAll(container *container.Container, router chi.Router) error {
	modulesDir, err := registry.ModulesDir()
	if err != nil {
		return err
	}
	return registry.AutoRegister(container, modulesDir, router)
}
