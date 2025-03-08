package services

import (
	"context"
	"github.com/mrkbwp/gotube/internal/domain/entity"
	"github.com/mrkbwp/gotube/internal/domain/repositories"
	"github.com/mrkbwp/gotube/internal/infrastructure/conversion"
	"github.com/mrkbwp/gotube/internal/infrastructure/storage"
)

type ConversionService struct {
	queue  *conversion.ConversionQueue
	ffmpeg *conversion.FFmpegService
}

func NewConversionService(
	videoRepo repositories.VideoRepository,
	storageClient *storage.MinioClient,
	tempDir string,
) *ConversionService {
	ffmpeg := conversion.NewFFmpegService(tempDir)

	service := &ConversionService{
		ffmpeg: ffmpeg,
	}

	queue := conversion.NewConversionQueue(videoRepo, storageClient, ffmpeg)
	service.queue = queue

	return service
}

func (s *ConversionService) StartConversionQueue() {
	s.queue.Start()
}

func (s *ConversionService) StopConversionQueue() {
	s.queue.Stop()
}

func (s *ConversionService) ConvertVideo(ctx context.Context, video *entity.Video, quality *entity.VideoQuality) error {
	return s.queue.ConvertVideo(ctx, video, quality)
}
