package models

import (
	"time"

	"gorm.io/gorm"
)

// MusicVideo model sesuai struktur tabel
type MusicVideo struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	ArtistID    int            `json:"artist_id" gorm:"not null;index"`
	Artist      string         `json:"artist" gorm:"size:255;not null"`
	ReleaseDate *time.Time     `json:"release_date" gorm:"type:date"`
	Duration    string         `json:"duration" gorm:"size:10;not null"` // Format: MM:SS atau HH:MM:SS
	Genre       string         `json:"genre" gorm:"size:100;not null"`
	Description *string        `json:"description" gorm:"type:text"`
	VideoURL     string         `json:"video_url" gorm:"size:500;not null"`
	Thumbnail    *string        `json:"thumbnail" gorm:"size:255"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (MusicVideo) TableName() string {
	return "music_videos"
}

