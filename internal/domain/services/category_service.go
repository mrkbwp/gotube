package services

import (
	"context"
	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// CategoryService определяет интерфейс для бизнес-логики категорий
type CategoryService interface {
	// GetAllCategories возвращает все категории
	GetAllCategories(ctx context.Context) ([]*entity.Category, error)

	// GetCategoryByID возвращает категорию по ID
	GetCategoryByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)

	// CreateCategory создает новую категорию
	CreateCategory(ctx context.Context, category *entity.Category) error

	// UpdateCategory обновляет категорию
	UpdateCategory(ctx context.Context, category *entity.Category) error

	// DeleteCategory удаляет категорию
	DeleteCategory(ctx context.Context, id uuid.UUID) error
}
