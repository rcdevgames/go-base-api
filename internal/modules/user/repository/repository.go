package repository

import (
	"context"
	"errors"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/entity"
)

// UserRepository describes data access behaviour for users.
type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
}

// ErrNotFound signals when a record is missing from the data source.
var ErrNotFound = errors.New("user not found")
