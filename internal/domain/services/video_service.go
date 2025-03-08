package services

import (
	"context"
	"github.com/google/uuid"
	"io"

	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// VideoService определяет интерфейс для бизнес-логики видео
type VideoService interface {
	// UploadVideo загружает новое видео
	UploadVideo(ctx context.Context, userID, categoryID uuid.UUID, file io.Reader, filename, title, description string) (*entity.Video, error)

	// GetVideoByCode возвращает информацию о видео по коду
	GetVideoByCode(ctx context.Context, code string) (*entity.Video, error)

	// GetVideoByID возвращает информацию о видео по ID
	GetVideoByID(ctx context.Context, id uuid.UUID) (*entity.Video, error)

	// GetNewVideos возвращает список новых видео с пагинацией
	GetNewVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error)

	// GetPopularVideos возвращает список популярных видео с пагинацией
	GetPopularVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error)

	// GetUserVideos возвращает видео пользователя с пагинацией
	GetUserVideos(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Video, int64, error)

	// UpdateVideo обновляет информацию о видео
	UpdateVideo(ctx context.Context, id, categoryID uuid.UUID, title, description string) (*entity.Video, error)

	// DeleteVideo удаляет видео
	DeleteVideo(ctx context.Context, id uuid.UUID) error

	// LikeVideo добавляет лайк видео
	LikeVideo(ctx context.Context, videoID, userID uuid.UUID) error

	// DislikeVideo добавляет дизлайк видео
	DislikeVideo(ctx context.Context, videoID, userID uuid.UUID) error

	// ViewVideo регистрирует просмотр видео
	ViewVideo(ctx context.Context, videoId uuid.UUID, userID *uuid.UUID, userIp string) error

	// GetVideoFiles получение видео файлов
	GetVideoFiles(ctx context.Context, video *entity.Video) ([]*entity.VideoFile, error)
}
