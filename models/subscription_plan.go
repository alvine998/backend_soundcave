package models

import (
	"time"

	"gorm.io/gorm"
)

// SubscriptionPlan model sesuai struktur tabel
type SubscriptionPlan struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name         string         `json:"name" gorm:"size:255;not null"`
	Price        string         `json:"price" gorm:"size:50;not null"`
	Duration     string         `json:"duration" gorm:"size:50;not null"`
	Features     JSONB          `json:"features" gorm:"type:json;not null"`
	MaxDownloads int            `json:"max_downloads" gorm:"default:-1"` // -1 means unlimited
	MaxPlaylists int            `json:"max_playlists" gorm:"default:-1"`   // -1 means unlimited
	AudioQuality string         `json:"audio_quality" gorm:"size:50;not null"`
	AdsEnabled   bool           `json:"ads_enabled" gorm:"type:tinyint(1);default:1"`
	OfflineMode  bool           `json:"offline_mode" gorm:"type:tinyint(1);default:0"`
	IsPopular    *bool          `json:"is_popular" gorm:"type:tinyint(1);default:0"`
	Description  string         `json:"description" gorm:"type:text;not null"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (SubscriptionPlan) TableName() string {
	return "subscription_plans"
}

