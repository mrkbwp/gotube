package conversion

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/infrastructure/storage"
	"github.com/mrkbwp/gotube/pkg/constants"
)

type ConversionQueue struct {
	videoRepo     repositories.VideoRepository
	storageClient *storage.MinioClient
	ffmpeg        *FFmpegService

	activeConversions sync.Map
	ticker            *time.Ticker
	stopChan          chan struct{}
}

func NewConversionQueue(
	videoRepo repositories.VideoRepository,
	storageClient *storage.MinioClient,
	ffmpeg *FFmpegService,
) *ConversionQueue {
	log.Println("Initializing conversion queue")
	return &ConversionQueue{
		videoRepo:     videoRepo,
		storageClient: storageClient,
		ffmpeg:        ffmpeg,
		ticker:        time.NewTicker(constants.ConversionCheckInterval),
		stopChan:      make(chan struct{}),
	}
}

func (q *ConversionQueue) Start() {
	log.Println("Starting conversion queue")
	go q.processQueue()
}

func (q *ConversionQueue) Stop() {
	log.Println("Stopping conversion queue")
	q.ticker.Stop()
	close(q.stopChan)
}

func (q *ConversionQueue) processQueue() {
	log.Println("Starting conversion queue processor")
	for {
		select {
		case <-q.ticker.C:
			count := 0
			q.activeConversions.Range(func(k, _ interface{}) bool {
				count++
				log.Printf("Active conversion: %v", k)
				return true
			})
			log.Printf("Current active conversions: %d", count)

			if count >= constants.MaxConcurrentConversions {
				log.Printf("Max concurrent conversions reached (%d), skipping", constants.MaxConcurrentConversions)
				continue
			}

			ctx := context.Background()
			videos, err := q.videoRepo.GetVideosForConversion(ctx, constants.MaxConcurrentConversions-count)
			if err != nil {
				log.Printf("Failed to get videos for conversion: %v", err)
				continue
			}
			log.Printf("Found %d videos for conversion", len(videos))

			for _, video := range videos {
				if _, exists := q.activeConversions.Load(video.ID); exists {
					log.Printf("Video %s is already being converted, skipping", video.ID)
					continue
				}

				log.Printf("Starting conversion for video %s", video.ID)
				if err := q.videoRepo.UpdateStatus(ctx, video.ID, string(constants.VideoStatusProcessing)); err != nil {
					log.Printf("Failed to update video status: %v", err)
					continue
				}

				q.activeConversions.Store(video.ID, true)
				go func(v *entity.Video) {
					defer q.activeConversions.Delete(v.ID)
					log.Printf("Starting conversion goroutine for video %s", v.ID)

					if err := q.convertVideo(v); err != nil {
						log.Printf("Failed to convert video %s: %v", v.ID, err)
						_ = q.videoRepo.UpdateStatus(ctx, v.ID, string(constants.VideoStatusError))
					}
					log.Printf("Finished conversion goroutine for video %s", v.ID)
				}(video)
			}

		case <-q.stopChan:
			log.Println("Received stop signal, stopping queue processor")
			return
		}
	}
}

func (q *ConversionQueue) convertVideo(video *entity.Video) error {
	ctx := context.Background()
	log.Printf("Starting video conversion for %s", video.ID)

	qualities, err := q.videoRepo.GetVideoQualities(ctx)
	if err != nil {
		log.Printf("Failed to get qualities for video %s: %v", video.ID, err)
		return fmt.Errorf("failed to get qualities: %w", err)
	}
	log.Printf("Got %d qualities for conversion", len(qualities))

	originalFilePath := video.GetStorageFilePath(constants.VideoQualityOriginal)
	inputFile := filepath.Join(q.ffmpeg.tempDir, video.Filename)
	log.Printf("Downloading original file from %s to %s", originalFilePath, inputFile)

	if err := q.storageClient.DownloadFile(ctx, video.BucketID, originalFilePath, inputFile); err != nil {
		log.Printf("Failed to download original file for video %s: %v", video.ID, err)
		return fmt.Errorf("failed to download original file: %w", err)
	}
	defer q.cleanupTempFile(inputFile)

	// Получаем длительность видео
	duration, err := q.ffmpeg.GetVideoInfo(inputFile)
	if err != nil {
		log.Printf("Failed to get video duration: %v", err)
		return fmt.Errorf("failed to get video duration: %w", err)
	}
	log.Printf("Video duration: %d seconds", duration)

	// Генерируем имя файла для thumbnail с правильным расширением
	thumbnailFilename := strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename)) + ".jpg"
	thumbnailPath := filepath.Join(q.ffmpeg.tempDir, thumbnailFilename)
	if err := q.ffmpeg.GenerateThumbnail(inputFile, thumbnailPath); err != nil {
		log.Printf("Failed to generate thumbnail: %v", err)
		return fmt.Errorf("failed to generate thumbnail: %w", err)
	}
	defer q.cleanupTempFile(thumbnailPath)

	// Открываем thumbnail для загрузки
	thumbFile, err := os.Open(thumbnailPath)
	if err != nil {
		log.Printf("Failed to open thumbnail file: %v", err)
		return fmt.Errorf("failed to open thumbnail: %w", err)
	}
	defer thumbFile.Close()

	// Генерируем путь для thumbnail без домена
	thumbnailStoragePath := fmt.Sprintf("%s/%s/%s/%s.jpg",
		video.ShardID,
		video.PathSegment1,
		video.PathSegment2,
		strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename)),
	)

	log.Printf("Thumbnail storage path: %s", thumbnailStoragePath)

	// Загружаем thumbnail в отдельный бакет
	if err := q.storageClient.UploadFile(ctx, constants.ThumbnailsBucket, thumbnailStoragePath, thumbFile); err != nil {
		log.Printf("Failed to upload thumbnail: %v", err)
		return fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Формируем прямой URL для thumbnail
	thumbnailURL := fmt.Sprintf("%s/%s",
		q.storageClient.GetPublicURL(constants.ThumbnailsBucket, ""), // base URL
		thumbnailStoragePath,
	)
	log.Printf("Storing thumbnail path: %s", thumbnailURL)

	// Обновляем thumbnail и duration
	if err := q.videoRepo.UpdateThumbnailAndDuration(ctx, video.ID, thumbnailURL, duration); err != nil {
		log.Printf("Failed to update video info: %v", err)
		return fmt.Errorf("failed to update video info: %w", err)
	}
	log.Printf("Updated video info with thumbnail path: %s and duration: %d", thumbnailURL, duration)

	lowQuality := qualities[0]
	if err := q.convertToQuality(ctx, video, lowQuality, inputFile); err != nil {
		return err
	}

	if err := q.videoRepo.UpdateStatus(ctx, video.ID, string(constants.VideoStatusReady)); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	for _, quality := range qualities[1:] {
		if err := q.convertToQuality(ctx, video, quality, inputFile); err != nil {
			log.Printf("Failed to convert to quality %s: %v", quality.Name, err)
			continue
		}
	}

	return nil
}

func (q *ConversionQueue) convertToQuality(ctx context.Context, video *entity.Video, quality *entity.VideoQuality, inputFile string) error {
	log.Printf("Starting conversion to quality %s for video %s", quality.Name, video.ID)

	// Генерируем имя выходного файла с качеством
	outputFilename := fmt.Sprintf("%s_%s%s",
		strings.TrimSuffix(video.Filename, filepath.Ext(video.Filename)),
		quality.Name,
		filepath.Ext(video.Filename),
	)
	outputFile := filepath.Join(q.ffmpeg.tempDir, outputFilename)
	log.Printf("Output file will be: %s", outputFile)

	log.Printf("Starting FFmpeg conversion for video %s to quality %s", video.ID, quality.Name)
	if err := q.ffmpeg.ConvertVideo(inputFile, outputFile, quality); err != nil {
		log.Printf("FFmpeg conversion failed for video %s quality %s: %v", video.ID, quality.Name, err)
		return fmt.Errorf("failed to convert video: %w", err)
	}
	defer q.cleanupTempFile(outputFile)

	file, err := os.Open(outputFile)
	if err != nil {
		log.Printf("Failed to open converted file %s: %v", outputFile, err)
		return fmt.Errorf("failed to open converted file: %w", err)
	}
	defer file.Close()

	storageFilePath := video.GetStorageFilePath(quality.Name)
	log.Printf("Uploading converted file for video %s quality %s to %s", video.ID, quality.Name, storageFilePath)
	if err := q.storageClient.UploadFile(ctx, video.BucketID, storageFilePath, file); err != nil {
		log.Printf("Failed to upload converted file for video %s quality %s: %v", video.ID, quality.Name, err)
		return fmt.Errorf("failed to upload converted file: %w", err)
	}

	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		log.Printf("Failed to get file info for %s: %v", outputFile, err)
		return fmt.Errorf("failed to get file info: %w", err)
	}
	log.Printf("Converted file size: %d bytes", fileInfo.Size())

	videoFile := &entity.VideoFile{
		VideoID:   video.ID,
		QualityID: quality.ID,
		Format:    filepath.Ext(video.Filename),
		FileSize:  fileInfo.Size(),
		Width:     quality.Width,
		Height:    quality.Height,
		Bitrate:   quality.Bitrate,
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	log.Printf("Creating video file record for video %s quality %s", video.ID, quality.Name)
	if err := q.videoRepo.CreateVideoFile(ctx, videoFile); err != nil {
		log.Printf("Failed to create video file record for video %s quality %s: %v", video.ID, quality.Name, err)
		return fmt.Errorf("failed to create video file record: %w", err)
	}

	log.Printf("Successfully completed conversion to quality %s for video %s", quality.Name, video.ID)
	return nil
}

func (q *ConversionQueue) cleanupTempFile(filename string) {
	log.Printf("Cleaning up temp file: %s", filename)
	if err := os.Remove(filename); err != nil {
		log.Printf("Failed to remove temp file %s: %v", filename, err)
	}
}

func (q *ConversionQueue) ConvertVideo(ctx context.Context, video *entity.Video, quality *entity.VideoQuality) error {
	log.Printf("Manual conversion requested for video %s to quality %s", video.ID, quality.Name)

	if _, exists := q.activeConversions.Load(video.ID); exists {
		log.Printf("Video %s is already being converted", video.ID)
		return fmt.Errorf("video is already being converted")
	}

	q.activeConversions.Store(video.ID, true)
	defer q.activeConversions.Delete(video.ID)

	originalFilePath := video.GetStorageFilePath(constants.VideoQualityOriginal)
	inputFile := filepath.Join(q.ffmpeg.tempDir, video.Filename)

	log.Printf("Downloading original file for manual conversion from %s to %s", originalFilePath, inputFile)
	if err := q.storageClient.DownloadFile(ctx, video.BucketID, originalFilePath, inputFile); err != nil {
		log.Printf("Failed to download original file for manual conversion: %v", err)
		return fmt.Errorf("failed to download original file: %w", err)
	}
	defer q.cleanupTempFile(inputFile)

	return q.convertToQuality(ctx, video, quality, inputFile)
}
