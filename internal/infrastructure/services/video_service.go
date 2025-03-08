package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mrkbwp/gotube/pkg/constants"
	"github.com/mrkbwp/gotube/pkg/kafka"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/domain/services"
	"github.com/redis/go-redis/v9"

	"github.com/mrkbwp/gotube/internal/infrastructure/storage"
)

// VideoService реализует интерфейс VideoService
type VideoService struct {
	videoRepo     repositories.VideoRepository
	storageClient *storage.MinioClient
	kafkaProducer *kafka.Producer
	redisClient   *redis.Client
	shardCount    int
}

// NewVideoService создает новый экземпляр VideoService
func NewVideoService(
	videoRepo repositories.VideoRepository,
	storageClient *storage.MinioClient,
	kafkaProducer *kafka.Producer,
	redisClient *redis.Client,
	shardCount int,
) services.VideoService {
	return &VideoService{
		videoRepo:     videoRepo,
		storageClient: storageClient,
		kafkaProducer: kafkaProducer,
		redisClient:   redisClient,
		shardCount:    shardCount,
	}
}

// UploadVideo загружает новое видео
func (s *VideoService) UploadVideo(
	ctx context.Context,
	userID, categoryID uuid.UUID,
	file io.Reader,
	originalFilename, title, description string,
) (*entity.Video, error) {
	// Генерируем уникальный код для видео
	var videoCode string
	// Проверка уникальности генерации кода
	for {
		videoCodeGen, err := s.generateUniquePublicID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate video code: %w", err)
		}
		_, err = s.videoRepo.GetByCode(ctx, videoCodeGen)
		if err != nil {
			if errors.Is(err, constants.ErrNotFound) {
				videoCode = videoCodeGen
				break
			}
			return nil, fmt.Errorf("failed to generate video code: %w", err)
		}
	}

	filename := generateFilename(originalFilename)

	// Генерируем пути для хранения
	bucketID, shardID, segment1, segment2 := s.generateStoragePath(filename)

	var video *entity.Video
	// Создаем запись о видео
	video = &entity.Video{
		ID:          uuid.New(),
		VideoCode:   videoCode,
		UserID:      userID,
		Title:       title,
		Description: description,
		CategoryID:  categoryID,

		BucketID:         bucketID,
		ShardID:          shardID,
		PathSegment1:     segment1,
		PathSegment2:     segment2,
		Filename:         filename,
		OriginalFilename: originalFilename,

		Status:    string(constants.VideoStatusUploaded),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	storagePath := video.GetStoragePath(constants.VideoQualityOriginal)

	// Сохраняем файл в хранилище
	objectName := filepath.Join(storagePath, filename)
	if err := s.storageClient.UploadFile(ctx, bucketID, objectName, file); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	// Сохраняем метаданные в БД
	if err := s.videoRepo.Create(ctx, video); err != nil {
		// В случае ошибки удаляем загруженный файл
		_ = s.storageClient.DeleteFile(ctx, bucketID, objectName)
		return nil, fmt.Errorf("failed to create video record: %w", err)
	}

	// Отправляем сообщение в Kafka для обработки
	message := kafka.VideoProcessingMessage{
		VideoID:      video.ID,
		BucketID:     bucketID,
		ShardID:      shardID,
		PathSegment1: segment1,
		PathSegment2: segment2,
		Filename:     filename,
	}

	if err := s.kafkaProducer.SendVideoProcessingMessage(ctx, message); err != nil {
		// Логируем ошибку, но не отменяем загрузку
		fmt.Printf("Failed to send processing message: %v\n", err)
	}

	return video, nil
}

// GetVideoByID возвращает информацию о видео по ID
func (s *VideoService) GetVideoByID(ctx context.Context, id uuid.UUID) (*entity.Video, error) {
	// Если нет в кеше, получаем из БД
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Проверяем статус и доступность
	//if err := s.validateVideoAccess(video); err != nil {
	//	return nil, err
	//}

	return video, nil
}

// GetVideoByCode возвращает информацию о видео по коду
func (s *VideoService) GetVideoByCode(ctx context.Context, code string) (*entity.Video, error) {
	// Если нет в кеше, получаем из БД
	video, err := s.videoRepo.GetByCode(ctx, code)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	// Проверяем статус и доступность
	//if err := s.validateVideoAccess(video); err != nil {
	//	return nil, err
	//}

	return video, nil
}

func (s *VideoService) GetVideoFiles(ctx context.Context, video *entity.Video) ([]*entity.VideoFile, error) {
	files, err := s.videoRepo.GetVideoFiles(ctx, video.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video files: %w", err)
	}

	// Для каждого файла генерируем временную ссылку
	for _, file := range files {
		url, err := s.storageClient.GetFileURL(
			ctx,
			video.BucketID,
			video.GetStorageFilePath(file.QualityName),
			int(time.Minute*10),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to generate url: %w", err)
		}

		file.URL = url
	}

	// Сортируем файлы по качеству (от низкого к высокому)
	sort.Slice(files, func(i, j int) bool {
		// Преобразуем качество в числа для сравнения (240p -> 240)
		qi := strings.TrimSuffix(files[i].QualityName, "p")
		qj := strings.TrimSuffix(files[j].QualityName, "p")

		vi, _ := strconv.Atoi(qi)
		vj, _ := strconv.Atoi(qj)

		return vi < vj
	})

	return files, nil
}

// GetNewVideos возвращает список новых видео с пагинацией
func (s *VideoService) GetNewVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error) {
	if err := s.validatePagination(page, limit); err != nil {
		return nil, 0, err
	}

	videos, total, err := s.videoRepo.GetNewVideos(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get new videos: %w", err)
	}

	return videos, total, nil
}

// GetPopularVideos возвращает список популярных видео с пагинацией
func (s *VideoService) GetPopularVideos(ctx context.Context, page, limit int) ([]*entity.Video, int64, error) {
	if err := s.validatePagination(page, limit); err != nil {
		return nil, 0, err
	}

	videos, total, err := s.videoRepo.GetPopularVideos(ctx, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get popular videos: %w", err)
	}

	return videos, total, nil
}

// GetUserVideos возвращает видео пользователя с пагинацией
func (s *VideoService) GetUserVideos(ctx context.Context, userID uuid.UUID, page, limit int) ([]*entity.Video, int64, error) {
	if err := s.validatePagination(page, limit); err != nil {
		return nil, 0, err
	}

	videos, total, err := s.videoRepo.GetUserVideos(ctx, userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user videos: %w", err)
	}

	return videos, total, nil
}

// UpdateVideo обновляет информацию о видео
func (s *VideoService) UpdateVideo(ctx context.Context, id, categoryID uuid.UUID, title, description string) (*entity.Video, error) {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return nil, constants.ErrVideoNotFound
		}
		return nil, fmt.Errorf("failed to get video: %w", err)
	}

	video.Title = title
	video.Description = description
	video.CategoryID = categoryID

	if err := s.videoRepo.Update(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to update video: %w", err)
	}

	return video, nil
}

// DeleteVideo удаляет видео
func (s *VideoService) DeleteVideo(ctx context.Context, id uuid.UUID) error {
	video, err := s.videoRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrVideoNotFound
		}
		return fmt.Errorf("failed to get video: %w", err)
	}

	if err := s.videoRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete video: %w", err)
	}

	// Удаляем файл из хранилища
	objectName := filepath.Join(
		video.PathSegment1,
		video.PathSegment2,
		video.Filename,
	)
	if err := s.storageClient.DeleteFile(ctx, video.BucketID, objectName); err != nil {
		fmt.Printf("Failed to delete file: %v\n", err)
	}

	return nil
}

// LikeVideo добавляет лайк видео
func (s *VideoService) LikeVideo(ctx context.Context, videoID, userID uuid.UUID) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrVideoNotFound
		}
		return fmt.Errorf("failed to get video: %w", err)
	}

	reaction, err := s.videoRepo.GetUserReaction(ctx, videoID, userID)
	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		return fmt.Errorf("failed to get reaction: %w", err)
	}

	if reaction != nil && reaction.Type == "like" {
		return s.videoRepo.DeleteReaction(ctx, videoID, userID)
	}

	newReaction := &entity.VideoReaction{
		VideoID:   video.ID,
		UserID:    userID,
		Type:      "like",
		CreatedAt: time.Now(),
	}

	if reaction != nil {
		err = s.videoRepo.UpdateReaction(ctx, newReaction)
	} else {
		err = s.videoRepo.CreateReaction(ctx, newReaction)
	}

	if err != nil {
		return fmt.Errorf("failed to save reaction: %w", err)
	}

	return nil
}

// DislikeVideo добавляет дизлайк видео
func (s *VideoService) DislikeVideo(ctx context.Context, videoID, userID uuid.UUID) error {
	video, err := s.videoRepo.GetByID(ctx, videoID)
	if err != nil {
		if errors.Is(err, constants.ErrNotFound) {
			return constants.ErrVideoNotFound
		}
		return fmt.Errorf("failed to get video: %w", err)
	}

	reaction, err := s.videoRepo.GetUserReaction(ctx, videoID, userID)
	if err != nil && !errors.Is(err, constants.ErrNotFound) {
		return fmt.Errorf("failed to get reaction: %w", err)
	}

	if reaction != nil && reaction.Type == "dislike" {
		return s.videoRepo.DeleteReaction(ctx, videoID, userID)
	}

	newReaction := &entity.VideoReaction{
		VideoID:   video.ID,
		UserID:    userID,
		Type:      "dislike",
		CreatedAt: time.Now(),
	}

	if reaction != nil {
		err = s.videoRepo.UpdateReaction(ctx, newReaction)
	} else {
		err = s.videoRepo.CreateReaction(ctx, newReaction)
	}

	if err != nil {
		return fmt.Errorf("failed to save reaction: %w", err)
	}

	return nil
}

// ViewVideo регистрирует просмотр видео
func (s *VideoService) ViewVideo(ctx context.Context, videoID uuid.UUID, viewerID *uuid.UUID, userIP string) error {
	if s.redisClient == nil {
		return s.videoRepo.IncrementViews(ctx, videoID)
	}

	viewKey := fmt.Sprintf("video:view:%s:%s", videoID, viewerID)

	exists, err := s.redisClient.Exists(ctx, viewKey).Result()
	if err != nil {
		return fmt.Errorf("failed to check view existence: %w", err)
	}

	if exists == 1 {
		return nil
	}

	if err := s.videoRepo.IncrementViews(ctx, videoID); err != nil {
		return fmt.Errorf("failed to increment views: %w", err)
	}

	if err := s.redisClient.Set(ctx, viewKey, "1", constants.VideoViewTimeout).Err(); err != nil {
		fmt.Printf("Failed to save view info: %v\n", err)
	}

	return nil
}

// Вспомогательные методы

func (s *VideoService) validateVideoAccess(video *entity.Video) error {
	if video.IsBlocked {
		return constants.ErrVideoBlocked
	}

	if video.IsPrivate {
		return constants.ErrVideoPrivate
	}

	if video.Status != string(constants.VideoStatusReady) {
		if video.Status == string(constants.VideoStatusUploaded) || video.Status == string(constants.VideoStatusProcessing) {
			return constants.ErrVideoProcessing
		}
		return constants.ErrInvalidStatus
	}

	return nil
}

func (s *VideoService) validatePagination(page, limit int) error {
	if page < 1 {
		return constants.ErrInvalidPagination
	}

	if limit < 1 {
		limit = constants.DefaultLimit
	}

	if limit > constants.MaxLimit {
		limit = constants.MaxLimit
	}

	return nil
}

func (s *VideoService) generateUniquePublicID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	id := base64.URLEncoding.EncodeToString(b)
	id = strings.ReplaceAll(id, "+", "-")
	id = strings.ReplaceAll(id, "/", "_")

	if len(id) > 11 {
		id = id[:11]
	}

	return id, nil
}

func generateFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	timestamp := time.Now().UnixNano()
	randomStr := make([]byte, 8)
	rand.Read(randomStr)
	return fmt.Sprintf("%d_%x%s", timestamp, randomStr, ext)
}

func (s *VideoService) generateStoragePath(filename string) (string, string, string, string) {
	segment1 := filename[:2]
	segment2 := filename[2:4]
	shardID := strconv.Itoa(int(filename[0]) % s.shardCount)
	bucketID := constants.VideoBucket

	return bucketID, shardID, segment1, segment2
}
