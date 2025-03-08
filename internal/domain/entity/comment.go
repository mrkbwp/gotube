package entity

import (
	"github.com/google/uuid"
	"time"
)

// Comment представляет комментарий к видео
type Comment struct {
	ID        uuid.UUID `json:"id" db:"id"`
	VideoID   uuid.UUID `json:"video_id" db:"video_id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Text      string    `json:"text" db:"text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	// Для построения дерева комментариев
	ParentID *uuid.UUID `json:"parent_id,omitempty" db:"parent_id"`
}
