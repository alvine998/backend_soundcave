package models

import (
	"time"

	"gorm.io/gorm"
)

// Playlist model sesuai struktur tabel
type Playlist struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      uint           `json:"user_id" gorm:"not null;index"`
	Name        string         `json:"name" gorm:"size:255;not null"`
	Description *string        `json:"description" gorm:"type:text"`
	IsPublic    *bool          `json:"is_public" gorm:"type:tinyint(1);default:1"`
	CoverImage  *string        `json:"cover_image" gorm:"size:255"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Playlist) TableName() string {
	return "playlists"
}
