package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/entity"
)

type stubRepo struct {
	user *entity.User
	err  error
}

func (s stubRepo) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func (s stubRepo) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		user := &entity.User{ID: 1, Name: "Test", Email: "test@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		service := NewService(stubRepo{user: user})

		result, err := service.GetByID(ctx, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.ID != user.ID {
			t.Fatalf("expected user ID %d, got %d", user.ID, result.ID)
		}
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := errors.New("boom")
		service := NewService(stubRepo{err: expectedErr})

		_, err := service.GetByID(ctx, 1)
		if !errors.Is(err, expectedErr) {
			t.Fatalf("expected error %v, got %v", expectedErr, err)
		}
	})
}
