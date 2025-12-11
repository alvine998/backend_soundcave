package models

import (
	"time"

	"gorm.io/gorm"
)

// News model sesuai struktur tabel
type News struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	Content     string         `json:"content" gorm:"type:text;not null"`
	Summary     *string        `json:"summary" gorm:"type:text"`
	Author      string         `json:"author" gorm:"size:255;not null"`
	Category    string         `json:"category" gorm:"size:100;not null"`
	ImageURL    *string        `json:"image_url" gorm:"size:500"`
	PublishedAt *time.Time     `json:"published_at" gorm:"type:datetime"`
	IsPublished *bool          `json:"is_published" gorm:"type:tinyint(1);default:0"`
	IsHeadline  *bool          `json:"is_headline" gorm:"type:tinyint(1);default:0"`
	Views       *int           `json:"views" gorm:"default:0"`
	Tags        *string        `json:"tags" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (News) TableName() string {
	return "news"
}
