package repositories

import (
	"context"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	// Create создает нового пользователя
	Create(ctx context.Context, user *entity.User) error

	// GetByID возвращает пользователя по ID
	GetByID(ctx context.Context, id string) (*entity.User, error)

	// GetByEmail возвращает пользователя по email
	GetByEmail(ctx context.Context, email string) (*entity.User, error)

	// Update обновляет данные пользователя
	Update(ctx context.Context, user *entity.User) error

	// UpdatePassword обновляет пароль пользователя
	UpdatePassword(ctx context.Context, id, passwordHash string) error

	// Delete удаляет пользователя
	Delete(ctx context.Context, id string) error
}
