package entity

import (
	"github.com/google/uuid"
	"time"
)

type VideoFile struct {
	ID          uuid.UUID `json:"id" db:"id,noi"`
	VideoID     uuid.UUID `json:"video_id" db:"video_id"`
	QualityID   uuid.UUID `json:"quality_id" db:"quality_id"`
	QualityName string    `json:"quality" db:"quality_name"`
	Format      string    `json:"format" db:"file_format"`
	URL         string    `json:"url" db:"-"`
	FileSize    int64     `json:"file_size" db:"file_size"`
	Width       int       `json:"width" db:"width"`
	Height      int       `json:"height" db:"height"`
	Bitrate     int       `json:"bitrate" db:"bitrate"`
	Status      string    `json:"-" db:"status"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
