package repositories

import (
	"context"
	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// TokenRepository определяет интерфейс для работы с refresh токенами
type TokenRepository interface {
	// Create создает новый токен
	Create(ctx context.Context, token *entity.Token) error

	// GetByRefreshToken находит токен по значению refresh токена
	GetByRefreshToken(ctx context.Context, refreshToken string) (*entity.Token, error)

	// DeleteByUserID удаляет все токены пользователя
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error

	// DeleteByID удаляет токен по ID
	DeleteByID(ctx context.Context, tokenID uuid.UUID) error

	// UpdateBlockStatus обновляет статус блокировки токена
	UpdateBlockStatus(ctx context.Context, tokenID uuid.UUID, isBlocked bool) error
}
