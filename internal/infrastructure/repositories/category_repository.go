package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/pkg/constants"

	"github.com/jmoiron/sqlx"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
)

// CategoryRepository реализует интерфейс CategoryRepository
type CategoryRepository struct {
	db *sqlx.DB
}

// NewCategoryRepository создает новый экземпляр CategoryRepository
func NewCategoryRepository(db *sqlx.DB) repositories.CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// GetAll возвращает все категории
func (r *CategoryRepository) GetAll(ctx context.Context) ([]*entity.Category, error) {
	query := `
		SELECT id, name, description, icon
		FROM categories
		ORDER BY name
	`

	var categories []*entity.Category
	err := r.db.SelectContext(ctx, &categories, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	return categories, nil
}

// GetByID возвращает категорию по ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	query := `
		SELECT id, name, description, icon
		FROM categories
		WHERE id = $1
	`

	var category entity.Category
	err := r.db.GetContext(ctx, &category, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &category, nil
}

// Create создает новую категорию
func (r *CategoryRepository) Create(ctx context.Context, category *entity.Category) error {
	query := `
		INSERT INTO categories (id, name, description, icon)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		category.ID, category.Name, category.Description, category.Icon,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// Update обновляет категорию
func (r *CategoryRepository) Update(ctx context.Context, category *entity.Category) error {
	query := `
		UPDATE categories
		SET name = $2, description = $3, icon = $4
		WHERE id = $1
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		category.ID, category.Name, category.Description, category.Icon,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	return nil
}

// Delete удаляет категорию
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
