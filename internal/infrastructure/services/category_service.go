package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"
	"github.com/redis/go-redis/v9"

	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/domain/services"
)

// CategoryService реализует интерфейс CategoryService
type CategoryService struct {
	categoryRepo repositories.CategoryRepository
	redisClient  *redis.Client
}

// NewCategoryService создает новый экземпляр CategoryService
func NewCategoryService(categoryRepo repositories.CategoryRepository, redisClient *redis.Client) services.CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		redisClient:  redisClient,
	}
}

// GetAllCategories возвращает все категории
func (s *CategoryService) GetAllCategories(ctx context.Context) ([]*entity.Category, error) {

	cacheKey := fmt.Sprintf("categories:all")

	categories, err := s.getCategoriesFromCache(ctx, cacheKey)
	if err == nil {
		return categories, nil
	}

	categories, err = s.categoryRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Сохраняем в кеш
	if err := s.cacheCategories(ctx, cacheKey, categories); err != nil {
		fmt.Printf("Failed to cache video: %v\n", err)
	}

	return categories, nil
}

// GetCategoryByID возвращает категорию по ID
func (s *CategoryService) GetCategoryByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

// CreateCategory создает новую категорию
func (s *CategoryService) CreateCategory(ctx context.Context, category *entity.Category) error {
	// Генерируем ID для новой категории, если его нет
	category.ID = uuid.New()
	return s.categoryRepo.Create(ctx, category)
}

// UpdateCategory обновляет категорию
func (s *CategoryService) UpdateCategory(ctx context.Context, category *entity.Category) error {
	// Проверяем, что категория существует
	_, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	return s.categoryRepo.Update(ctx, category)
}

// DeleteCategory удаляет категорию
func (s *CategoryService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	// Проверяем, что категория существует
	_, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	return s.categoryRepo.Delete(ctx, id)
}

func (s *CategoryService) getCategoriesFromCache(ctx context.Context, key string) ([]*entity.Category, error) {
	if s.redisClient == nil {
		return nil, errors.New("redis client not initialized")
	}

	data, err := s.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var categories []*entity.Category
	if err := json.Unmarshal(data, &categories); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *CategoryService) cacheCategories(ctx context.Context, key string, categories []*entity.Category) error {
	if s.redisClient == nil {
		return errors.New("redis client not initialized")
	}

	data, err := json.Marshal(categories)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, data, constants.CategoriesCacheDuration).Err()
}
