package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/entity"
	"github.com/rcdevgames/modular-monolith-clean/internal/modules/auth/repository"
	"github.com/rcdevgames/modular-monolith-clean/internal/server/security"
)

// Service exposes authentication use cases.
type Service interface {
	Login(ctx context.Context, email, password string) (*entity.TokenPair, error)
	Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error)
}

type serviceImpl struct {
	repo          repository.Repository
	accessSecret  string
	refreshSecret string
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewService builds a Service.
func NewService(repo repository.Repository, accessSecret string, accessTTL time.Duration, refreshSecret string, refreshTTL time.Duration) Service {
	return &serviceImpl{repo: repo, accessSecret: accessSecret, accessTTL: accessTTL, refreshSecret: refreshSecret, refreshTTL: refreshTTL}
}

func (s *serviceImpl) Login(ctx context.Context, email, password string) (*entity.TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if err := security.ComparePassword(user.PasswordHash, password); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return s.issueTokens(user)
}

func (s *serviceImpl) Refresh(ctx context.Context, refreshToken string) (*entity.TokenPair, error) {
	claims, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.refreshSecret), nil
	})
	if err != nil || !claims.Valid {
		return nil, errors.New("invalid refresh token")
	}
	mapClaims, ok := claims.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid refresh token claims")
	}
	if typ, _ := mapClaims["typ"].(string); typ != "refresh" {
		return nil, errors.New("invalid refresh token")
	}
	subject, _ := mapClaims["sub"].(string)
	role, _ := mapClaims["role"].(string)
	name, _ := mapClaims["name"].(string)
	uid := toInt64(mapClaims["uid"])
	user := &entity.User{ID: uid, Name: name, Email: subject, Role: role}
	return s.issueTokens(user)
}

func (s *serviceImpl) issueTokens(user *entity.User) (*entity.TokenPair, error) {
	accessClaims := jwt.MapClaims{
		"sub":  user.Email,
		"uid":  user.ID,
		"role": user.Role,
		"name": user.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.accessTTL).Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.accessSecret))
	if err != nil {
		return nil, err
	}
	refreshClaims := jwt.MapClaims{
		"sub":  user.Email,
		"uid":  user.ID,
		"role": user.Role,
		"name": user.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(s.refreshTTL).Unix(),
		"typ":  "refresh",
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.refreshSecret))
	if err != nil {
		return nil, err
	}
	return &entity.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func toInt64(value interface{}) int64 {
	switch v := value.(type) {
	case float64:
		return int64(v)
	case float32:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
