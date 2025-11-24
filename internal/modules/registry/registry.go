package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/go-chi/chi/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/server/container"
)

// Module exposes a unit capable of registering HTTP routes.
type Module interface {
	Name() string
	RegisterRoutes(r chi.Router)
}

// Builder creates a module instance with dependencies from the container.
type Builder func(*container.Container) (Module, error)

var (
	buildersMu sync.RWMutex
	builders   = map[string]Builder{}
)

// Register adds a module builder under a specific name.
func Register(name string, builder Builder) {
	buildersMu.Lock()
	defer buildersMu.Unlock()
	builders[name] = builder
}

// AutoRegister scans the modules directory and initializes each module whose builder exists.
func AutoRegister(container *container.Container, modulesDir string, router chi.Router) error {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return fmt.Errorf("read modules dir: %w", err)
	}
	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}
	sort.Strings(dirs)
	for _, dir := range dirs {
		builder, ok := getBuilder(dir)
		if !ok {
			continue
		}
		module, err := builder(container)
		if err != nil {
			return fmt.Errorf("build module %s: %w", dir, err)
		}
		module.RegisterRoutes(router)
	}
	return nil
}

func getBuilder(name string) (Builder, bool) {
	buildersMu.RLock()
	defer buildersMu.RUnlock()
	b, ok := builders[name]
	return b, ok
}

// ModulesDir returns the absolute path to the modules folder relative to the current working directory.
func ModulesDir() (string, error) {
	base, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "internal", "modules"), nil
}
