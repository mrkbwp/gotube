// entity/video.go
package entity

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"time"
)

type Metadata map[string]interface{}

// Value implements the driver.Valuer interface
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// Scan implements the sql.Scanner interface
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &m)
}

type Video struct {
	ID          uuid.UUID `json:"id" db:"id"`
	VideoCode   string    `json:"video_code" db:"video_code"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	CategoryID  uuid.UUID `json:"category_id" db:"category_id"`

	BucketID     string `json:"bucket_id" db:"bucket_id"`
	ShardID      string `json:"shard_id" db:"shard_id"`
	PathSegment1 string `json:"path_segment1" db:"path_segment1"`
	PathSegment2 string `json:"path_segment2" db:"path_segment2"`
	Filename     string `json:"filename" db:"filename"`

	ThumbnailURL string `json:"thumbnail_url" db:"thumbnail_url"`
	Duration     int    `json:"duration" db:"duration"`
	Views        int    `json:"views" db:"views"`
	Likes        int    `json:"likes" db:"likes"`
	Dislikes     int    `json:"dislikes" db:"dislikes"`
	Status       string `json:"status" db:"status"`

	IsBlocked    bool       `json:"is_blocked" db:"is_blocked"`
	IsPrivate    bool       `json:"is_private" db:"is_private"`
	ProcessedAt  *time.Time `json:"processed_at" db:"processed_at"`
	ErrorMessage *string    `json:"error_message" db:"error_message"`

	Metadata         Metadata `json:"metadata" db:"metadata"`
	OriginalFilename string   `json:"original_filename" db:"original_filename"`

	Files []*VideoFile `json:"video_files" db:"-"`

	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at,noi"`
}

// GetStoragePath возвращает полный путь к папке в хранилище
func (v *Video) GetStoragePath(quality string) string {
	return v.ShardID + "/" +
		v.PathSegment1 + "/" +
		v.PathSegment2 + "/" +
		quality
}

// GetStorageFilePath возвращает полный путь к видео в хранилище
func (v *Video) GetStorageFilePath(quality string) string {
	return v.GetStoragePath(quality) + "/" + v.Filename
}
