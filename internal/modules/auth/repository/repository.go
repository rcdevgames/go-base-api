package repository

import (
	"context"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/entity"
)

// Repository describes user lookup capabilities needed for auth.
type Repository interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}
