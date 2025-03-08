package entity

import (
	"github.com/google/uuid"
	"time"
)

type VideoQuality struct {
	ID        uuid.UUID `db:"id"`
	Name      string    `db:"name"`
	Height    int       `db:"height"`
	Width     int       `db:"width"`
	Bitrate   int       `db:"target_bitrate"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
