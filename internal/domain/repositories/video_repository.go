package repositories

import (
	"context"
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/internal/domain/entity"
)

// VideoRepository определяет интерфейс для работы с видео в базе данных
type VideoRepository interface {
	// Create создает новую запись о видео
	Create(ctx context.Context, video *entity.Video) error

	// GetByID возвращает видео по ID
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Video, error)

	// GetByCode возвращает видео по коду
	GetByCode(ctx context.Context, code string) (*entity.Video, error)

	// GetNewVideos возвращает список новых видео с пагинацией
	GetNewVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error)

	// GetPopularVideos возвращает список популярных видео с пагинацией
	GetPopularVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error)

	// GetUserVideos возвращает видео пользователя с пагинацией
	GetUserVideos(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Video, int64, error)

	// Update обновляет информацию о видео
	Update(ctx context.Context, video *entity.Video) error

	// Delete помечает видео как удаленное
	Delete(ctx context.Context, id uuid.UUID) error

	// IncrementViews увеличивает счетчик просмотров
	IncrementViews(ctx context.Context, id uuid.UUID) error

	// GetUserReaction возвращает реакцию пользователя на видео
	GetUserReaction(ctx context.Context, videoID, userID uuid.UUID) (*entity.VideoReaction, error)

	// CreateReaction создает реакцию пользователя на видео
	CreateReaction(ctx context.Context, reaction *entity.VideoReaction) error

	// UpdateReaction обновляет реакцию пользователя
	UpdateReaction(ctx context.Context, reaction *entity.VideoReaction) error

	// DeleteReaction удаляет реакцию пользователя
	DeleteReaction(ctx context.Context, videoID, userID uuid.UUID) error

	// GetVideoFiles получение файлов в разном качестве
	GetVideoFiles(ctx context.Context, videoID uuid.UUID) ([]*entity.VideoFile, error)

	// GetVideosForConversion получение видео для конвертации
	GetVideosForConversion(ctx context.Context, limit int) ([]*entity.Video, error)

	// GetVideoQualities получение списка качеств видео
	GetVideoQualities(ctx context.Context) ([]*entity.VideoQuality, error)

	// UpdateStatus обновление статуса видео
	UpdateStatus(ctx context.Context, videoID uuid.UUID, status string) error

	// CreateVideoFile добавление ссылки на видео в качестве
	CreateVideoFile(ctx context.Context, file *entity.VideoFile) error

	// UpdateThumbnailAndDuration обновляем картинку и длительность
	UpdateThumbnailAndDuration(ctx context.Context, videoID uuid.UUID, thumbnailURL string, duration int) error

	// RecalculateReactionCounts пересчет показателей
	RecalculateReactionCounts(ctx context.Context, videoID uuid.UUID) error
}
