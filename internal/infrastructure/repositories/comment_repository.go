package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/mrkbwp/gotube/pkg/constants"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
)

// CommentRepository реализует интерфейс CommentRepository
type CommentRepository struct {
	db *sqlx.DB
}

// NewCommentRepository создает новый экземпляр CommentRepository
func NewCommentRepository(db *sqlx.DB) repositories.CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

// Create создает новый комментарий
func (r *CommentRepository) Create(ctx context.Context, comment *entity.Comment) error {
	query := `
		INSERT INTO comments (id, video_id, user_id, parent_id, text, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	comment.ID = uuid.New()

	now := time.Now()
	if comment.CreatedAt.IsZero() {
		comment.CreatedAt = now
	}
	if comment.UpdatedAt.IsZero() {
		comment.UpdatedAt = now
	}

	var id string
	err := r.db.QueryRowContext(
		ctx,
		query,
		comment.ID, comment.VideoID, comment.UserID, comment.ParentID, comment.Text,
		comment.CreatedAt, comment.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID возвращает комментарий по ID
func (r *CommentRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Comment, error) {
	query := `
		SELECT id, video_id, user_id, parent_id, text, created_at, updated_at
		FROM comments
		WHERE id = $1
	`

	var comment entity.Comment
	err := r.db.GetContext(ctx, &comment, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, constants.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// GetVideoComments возвращает комментарии к видео с пагинацией
func (r *CommentRepository) GetVideoComments(ctx context.Context, videoID uuid.UUID, page, limit int) ([]*entity.Comment, int64, error) {
	query := `
		SELECT id, video_id, user_id, parent_id, text, created_at, updated_at
		FROM comments
		WHERE video_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	countQuery := `
		SELECT COUNT(*) FROM comments WHERE video_id = $1
	`

	// Вычисляем смещение для пагинации
	offset := (page - 1) * limit

	// Получаем комментарии
	var comments []*entity.Comment
	err := r.db.SelectContext(ctx, &comments, query, videoID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get comments: %w", err)
	}

	// Получаем общее количество
	var total int64
	err = r.db.GetContext(ctx, &total, countQuery, videoID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	return comments, total, nil
}

// Update обновляет комментарий
func (r *CommentRepository) Update(ctx context.Context, comment *entity.Comment) error {
	query := `
		UPDATE comments
		SET text = $2, updated_at = $3
		WHERE id = $1
	`

	comment.UpdatedAt = time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		comment.ID, comment.Text, comment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	return nil
}

// Delete удаляет комментарий
func (r *CommentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM comments
		WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	return nil
}
