package usecase

import (
	"context"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository"
)

// Service exposes user business capabilities.
type Service interface {
	GetByID(ctx context.Context, id int64) (*entity.User, error)
}

type serviceImpl struct {
	repo repository.UserRepository
}

// NewService creates a new user use case service.
func NewService(repo repository.UserRepository) Service {
	return &serviceImpl{repo: repo}
}

func (s *serviceImpl) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	return s.repo.GetByID(ctx, id)
}
