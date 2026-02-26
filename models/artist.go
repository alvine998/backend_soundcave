package models

import (
	"time"

	"gorm.io/gorm"
)

// Artist model sesuai struktur tabel
type Artist struct {
	ID            uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	RefUserID     *uint           `json:"ref_user_id" gorm:"index"` // Link to user ID if role is independent artist
	Name          string          `json:"name" gorm:"size:255;not null"`
	Bio           string          `json:"bio" gorm:"type:text;not null"`
	Genre         string          `json:"genre" gorm:"size:100"`
	Country       string          `json:"country" gorm:"size:100"`
	DebutYear     string          `json:"debut_year" gorm:"size:4"`
	Website       *string         `json:"website" gorm:"size:255"`
	Email         string          `json:"email" gorm:"size:255;not null"`
	Phone         *string         `json:"phone" gorm:"size:20"`
	SocialMedia   JSONB           `json:"social_media" gorm:"type:json"`
	ProfileImage  *string         `json:"profile_image" gorm:"column:profile_image;size:255"`
	Followers     JSONStringArray `json:"followers" gorm:"type:json"`
	TotalFollower int             `json:"total_follower" gorm:"default:0"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Artist) TableName() string {
	return "artists"
}
