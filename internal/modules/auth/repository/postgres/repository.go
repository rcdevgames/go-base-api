package postgres

import (
	"context"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/repository"
	userrepo "github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository"
)

type repositoryImpl struct {
	userRepo userrepo.UserRepository
}

// NewRepository wraps the user module repository for auth lookups.
func NewRepository(userRepo userrepo.UserRepository) repository.Repository {
	return &repositoryImpl{userRepo: userRepo}
}

func (r *repositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	user, err := r.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return &entity.User{
		ID:           user.ID,
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
	}, nil
}
