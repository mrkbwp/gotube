package repositories

import (
	"context"
	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// CategoryRepository определяет интерфейс для работы с категориями
type CategoryRepository interface {
	// GetAll возвращает все категории
	GetAll(ctx context.Context) ([]*entity.Category, error)

	// GetByID возвращает категорию по ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)

	// Create создает новую категорию
	Create(ctx context.Context, category *entity.Category) error

	// Update обновляет категорию
	Update(ctx context.Context, category *entity.Category) error

	// Delete удаляет категорию
	Delete(ctx context.Context, id uuid.UUID) error
}
