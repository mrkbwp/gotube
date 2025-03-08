package services

import (
	"context"
	"github.com/mrkbwp/gotube/internal/domain/entity"
)

type ConversionService interface {
	StartConversionQueue()
	StopConversionQueue()
	ConvertVideo(ctx context.Context, video *entity.Video, quality *entity.VideoQuality) error
}
