package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/repository"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/security"
)

type stubRepo struct {
	user *entity.User
	err  error
}

func (s stubRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestService_Login(t *testing.T) {
	repo := stubRepo{user: &entity.User{ID: 1, Name: "Tester", Email: "test@example.com", PasswordHash: hashPassword("secret"), Role: "user"}}
	service := NewService(repo, "access", time.Minute, "refresh", time.Hour)
	if _, err := service.Login(context.Background(), "test@example.com", "secret"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if _, err := service.Login(context.Background(), "test@example.com", "bad"); err == nil {
		t.Fatalf("expected error for invalid password")
	}
}

func TestService_Refresh(t *testing.T) {
	repo := stubRepo{user: &entity.User{ID: 1, Name: "Tester", Email: "test@example.com", PasswordHash: hashPassword("secret"), Role: "user"}}
	service := NewService(repo, "access", time.Minute, "refresh", time.Hour)
	pair, err := service.Login(context.Background(), "test@example.com", "secret")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if _, err := service.Refresh(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
}

// helper implementation for hashing using bcrypt.
func hashPassword(pwd string) string {
	hash, _ := security.HashPassword(pwd, 10)
	return hash
}

var _ repository.Repository = stubRepo{}
