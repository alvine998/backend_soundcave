package models

import (
	"time"

	"gorm.io/gorm"
)

// ArtistStreamStatus status stream
type ArtistStreamStatus string

const (
	StreamStatusScheduled ArtistStreamStatus = "scheduled"
	StreamStatusLive      ArtistStreamStatus = "live"
	StreamStatusEnded     ArtistStreamStatus = "ended"
)

// ArtistStream model untuk tracking live streaming artist
type ArtistStream struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	ArtistID    int32          `json:"artist_id" gorm:"type:int;not null;index"`
	Artist      Artist         `json:"artist" gorm:"foreignKey:ArtistID"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text"`
	Thumbnail   *string        `json:"thumbnail" gorm:"size:500"`
	ScheduledAt *time.Time     `json:"scheduled_at"`
	StreamKey   string         `json:"stream_key" gorm:"size:100;uniqueIndex"`
	IngestURL   string         `json:"ingest_url" gorm:"size:500"`
	PlaybackURL string         `json:"playback_url" gorm:"size:500"`
	Status      ArtistStreamStatus `json:"status" gorm:"type:enum('scheduled','live','ended');default:'scheduled'"`
	ViewerCount int            `json:"viewer_count" gorm:"default:0"`
	StartedAt   time.Time      `json:"started_at" gorm:"autoCreateTime"`
	EndedAt     *time.Time     `json:"ended_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (ArtistStream) TableName() string {
	return "artist_streams"
}
