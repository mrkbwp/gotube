package services

import (
	"context"
	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// CommentService определяет интерфейс для бизнес-логики комментариев
type CommentService interface {
	// AddComment добавляет новый комментарий
	AddComment(ctx context.Context, comment *entity.Comment) error

	// GetCommentByID возвращает комментарий по ID
	GetCommentByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error)

	// GetVideoComments возвращает комментарии к видео с пагинацией
	GetVideoComments(ctx context.Context, videoCode string, page, limit int) ([]*entity.Comment, int64, error)

	// UpdateComment обновляет комментарий
	UpdateComment(ctx context.Context, comment *entity.Comment) error

	// DeleteComment удаляет комментарий
	DeleteComment(ctx context.Context, id uuid.UUID) error
}
