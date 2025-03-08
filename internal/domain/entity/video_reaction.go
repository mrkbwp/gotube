package entity

import (
	"github.com/google/uuid"
	"time"
)

// VideoReaction представляет реакцию пользователя на видео (лайк/дислайк)
type VideoReaction struct {
	ID        uuid.UUID  `json:"id" db:"id,noi"`
	VideoID   uuid.UUID  `json:"video_id" db:"video_id"`
	UserID    uuid.UUID  `json:"user_id" db:"user_id"`
	Type      string     `json:"type" db:"type"` // "like" или "dislike"
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at,noi"`
}
