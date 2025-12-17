package models

import (
	"time"

	"gorm.io/gorm"
)

// CavelistStatus enum untuk status cavelist
type CavelistStatus string

const (
	CavelistStatusDraft   CavelistStatus = "draft"
	CavelistStatusPublish CavelistStatus = "publish"
)

// Cavelist model sesuai struktur tabel
type Cavelist struct {
	ID              uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title           string         `json:"title" gorm:"size:255;not null"`
	Description     *string        `json:"description" gorm:"type:text"`
	IsPromotion     *bool          `json:"is_promotion" gorm:"type:tinyint(1);default:0"`
	ExpiryPromotion *time.Time     `json:"expiry_promotion" gorm:"type:datetime"`
	Viewers         *int           `json:"viewers" gorm:"default:0"`
	Likes           *int           `json:"likes" gorm:"default:0"`
	Shares          *int           `json:"shares" gorm:"default:0"`
	VideoURL        string         `json:"video_url" gorm:"size:500;not null"`
	ArtistID        int            `json:"artist_id" gorm:"not null;index"`
	ArtistName      string         `json:"artist_name" gorm:"size:255;not null"`
	Status          CavelistStatus `json:"status" gorm:"type:enum('draft','publish');default:'draft'"`
	PublishedAt     *time.Time     `json:"published_at" gorm:"type:datetime"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Cavelist) TableName() string {
	return "cavelists"
}
