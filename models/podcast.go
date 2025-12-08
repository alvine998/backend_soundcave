package models

import (
	"time"

	"gorm.io/gorm"
)

// Podcast model sesuai struktur tabel
type Podcast struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title         string         `json:"title" gorm:"size:255;not null"`
	Host          string         `json:"host" gorm:"size:255;not null"`
	ReleaseDate   *time.Time     `json:"release_date" gorm:"type:date"`
	Duration      string         `json:"duration" gorm:"size:10;not null"` // Format: MM:SS atau HH:MM:SS
	Category      string         `json:"category" gorm:"size:100;not null"`
	Description   string         `json:"description" gorm:"type:text;not null"`
	EpisodeNumber *int           `json:"episode_number"`
	Season        *int           `json:"season"`
	VideoURL      string         `json:"video_url" gorm:"column:video_url;size:500;not null"`
	Thumbnail     *string        `json:"thumbnail" gorm:"size:255"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Podcast) TableName() string {
	return "podcasts"
}
