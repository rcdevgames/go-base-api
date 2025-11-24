package container

import (
	"database/sql"

	"github.com/rcdevgames/modular-monolith-clean/internal/config"
	"github.com/rcdevgames/modular-monolith-clean/internal/storage"
)

// Container exposes shared infrastructure dependencies to modules.
type Container struct {
	config *config.Config
	db     *sql.DB
	store  storage.Storage
}

// New builds a container instance.
func New(cfg *config.Config, db *sql.DB, store storage.Storage) *Container {
	return &Container{config: cfg, db: db, store: store}
}

// Config returns application configuration.
func (c *Container) Config() *config.Config {
	return c.config
}

// DB returns the primary SQL database handle.
func (c *Container) DB() *sql.DB {
	return c.db
}

// Storage returns the configured object storage backend.
func (c *Container) Storage() storage.Storage {
	return c.store
}
