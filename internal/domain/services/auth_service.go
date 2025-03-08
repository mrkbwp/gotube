package services

import (
	"context"
	"github.com/mrkbwp/gotube/pkg/jwt"

	"github.com/mrkbwp/gotube/internal/api/requests"
	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// AuthService определяет интерфейс для бизнес-логики аутентификации
type AuthService interface {
	// Register регистрирует нового пользователя
	Register(ctx context.Context, req requests.RegisterRequest) (*entity.User, *jwt.TokenPair, error)

	// Login авторизует пользователя
	Login(ctx context.Context, req requests.LoginRequest) (*entity.User, *jwt.TokenPair, error)

	// RefreshTokens обновляет пару токенов
	RefreshTokens(ctx context.Context, refreshToken string, userAgent string, clientIP string) (*jwt.TokenPair, error)

	// Logout выход из системы
	Logout(ctx context.Context, refreshToken string) error
}
