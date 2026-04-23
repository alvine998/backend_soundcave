package models

import (
	"time"

	"gorm.io/gorm"
)

// ArtistStreamStatus status stream
type ArtistStreamStatus string

const (
	StreamStatusPending   ArtistStreamStatus = "pending"   // created, waiting for SRS connection
	StreamStatusScheduled ArtistStreamStatus = "scheduled" // scheduled for a future time
	StreamStatusLive      ArtistStreamStatus = "live"      // SRS confirmed publish
	StreamStatusEnded     ArtistStreamStatus = "ended"     // SRS unpublished or manually ended
)

// ArtistStream model untuk tracking live streaming artist
type ArtistStream struct {
	ID          uint               `json:"id" gorm:"primaryKey;autoIncrement"`
	ArtistID    int32              `json:"artist_id" gorm:"type:int;not null;index"`
	Artist      Artist             `json:"artist" gorm:"foreignKey:ArtistID"`
	Title       string             `json:"title" gorm:"size:255;not null"`
	Description string             `json:"description" gorm:"type:text"`
	Thumbnail   *string            `json:"thumbnail" gorm:"size:500"`
	ScheduledAt *time.Time         `json:"scheduled_at"`
	StreamKey   string             `json:"stream_key" gorm:"size:100;uniqueIndex"`
	IngestURL   string             `json:"ingest_url" gorm:"size:500"`   // RTMP ingest for legacy
	WebRTCURL   string             `json:"webrtc_url" gorm:"size:500"`   // SRS WHIP endpoint for WebRTC push
	PlaybackURL string             `json:"playback_url" gorm:"size:500"` // HLS playback URL
	Status      ArtistStreamStatus `json:"status" gorm:"type:enum('pending','scheduled','live','ended');default:'pending'"`
	ViewerCount int                `json:"viewer_count" gorm:"default:0"`
	StartedAt   *time.Time         `json:"started_at"` // set when SRS confirms on_publish
	EndedAt     *time.Time         `json:"ended_at"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
	DeletedAt   gorm.DeletedAt     `json:"deleted_at" gorm:"index" swag:"-"`
}

// TableName mengembalikan nama tabel
func (ArtistStream) TableName() string {
	return "artist_streams"
}
