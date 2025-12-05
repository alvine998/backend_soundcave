package models

import (
	"time"

	"gorm.io/gorm"
)

// Music model sesuai struktur tabel
type Music struct {
	ID            uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title         string         `json:"title" gorm:"size:255;not null"`
	Artist        string         `json:"artist" gorm:"size:255;not null"`
	ArtistID      int            `json:"artist_id" gorm:"not null;index"`
	Album         *string        `json:"album" gorm:"size:255"`
	AlbumID       *int           `json:"album_id" gorm:"index"`
	Genre         string         `json:"genre" gorm:"size:100;not null"`
	ReleaseDate   *time.Time     `json:"release_date" gorm:"type:date"`
	Duration      string         `json:"duration" gorm:"size:10;not null"` // Format: MM:SS atau HH:MM:SS
	Language      string         `json:"language" gorm:"size:50;not null"`
	Explicit      *bool          `json:"explicit" gorm:"type:tinyint(1);default:0"`
	Lyrics        *string        `json:"lyrics" gorm:"type:text"`
	Description   *string        `json:"description" gorm:"type:text"`
	Tags          *string        `json:"tags" gorm:"type:text"`
	AudioFileURL  string         `json:"audio_file_url" gorm:"size:500;not null"`
	CoverImageURL *string        `json:"cover_image_url" gorm:"size:500"`
	PlayCount     *int           `json:"play_count" gorm:"default:0"`
	LikeCount     *int           `json:"like_count" gorm:"default:0"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Music) TableName() string {
	return "musics"
}

