package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"

	"github.com/google/uuid"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/domain/services"
)

// CommentService реализует интерфейс CommentService
type CommentService struct {
	commentRepo  repositories.CommentRepository
	videoService services.VideoService
}

// NewCommentService создает новый экземпляр CommentService
func NewCommentService(commentRepo repositories.CommentRepository, videoService services.VideoService) services.CommentService {
	return &CommentService{
		commentRepo:  commentRepo,
		videoService: videoService,
	}
}

// AddComment добавляет новый комментарий
func (s *CommentService) AddComment(ctx context.Context, comment *entity.Comment) error {
	// Генерируем ID для нового комментария, если его нет
	comment.ID = uuid.New()
	return s.commentRepo.Create(ctx, comment)
}

// GetCommentByID возвращает комментарий по ID
func (s *CommentService) GetCommentByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	return s.commentRepo.GetByID(ctx, id)
}

// GetVideoComments возвращает комментарии к видео с пагинацией
func (s *CommentService) GetVideoComments(ctx context.Context, videoCode string, page, limit int) ([]*entity.Comment, int64, error) {
	// Проверка входных параметров пагинации
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20 // Значение по умолчанию
	}

	videoInfo, err := s.videoService.GetVideoByCode(ctx, videoCode)
	if err != nil {
		return nil, 0, err
	}

	return s.commentRepo.GetVideoComments(ctx, videoInfo.ID, page, limit)
}

// UpdateComment обновляет комментарий
func (s *CommentService) UpdateComment(ctx context.Context, comment *entity.Comment) error {
	// Проверяем, что комментарий существует
	existingComment, err := s.commentRepo.GetByID(ctx, comment.ID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	// Сохраняем неизменяемые поля
	comment.UserID = existingComment.UserID
	comment.VideoID = existingComment.VideoID
	comment.CreatedAt = existingComment.CreatedAt
	comment.ParentID = existingComment.ParentID

	return s.commentRepo.Update(ctx, comment)
}

// DeleteComment удаляет комментарий
func (s *CommentService) DeleteComment(ctx context.Context, id uuid.UUID) error {
	// Проверяем, что комментарий существует
	_, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return fmt.Errorf("comment not found")
		}
		return fmt.Errorf("failed to get comment: %w", err)
	}

	return s.commentRepo.Delete(ctx, id)
}
