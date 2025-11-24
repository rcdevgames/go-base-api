package postgres

import (
	"context"
	"database/sql"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/user/repository"
)

type repositoryImpl struct {
	db *sql.DB
}

// NewRepository wires a PostgreSQL implementation for the UserRepository boundary.
func NewRepository(db *sql.DB) repository.UserRepository {
	return &repositoryImpl{db: db}
}

func (r *repositoryImpl) GetByID(ctx context.Context, id int64) (*entity.User, error) {
	const query = `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user entity.User
	if err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *repositoryImpl) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	const query = `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user entity.User
	if err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
