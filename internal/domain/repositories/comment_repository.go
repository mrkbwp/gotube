package repositories

import (
	"context"
	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// CommentRepository определяет интерфейс для работы с комментариями
type CommentRepository interface {
	// Create создает новый комментарий
	Create(ctx context.Context, comment *entity.Comment) error

	// GetByID возвращает комментарий по ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)

	// GetVideoComments возвращает комментарии к видео с пагинацией
	GetVideoComments(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*entity.Comment, int64, error)

	// Update обновляет комментарий
	Update(ctx context.Context, comment *entity.Comment) error

	// Delete удаляет комментарий
	Delete(ctx context.Context, id uuid.UUID) error
}
