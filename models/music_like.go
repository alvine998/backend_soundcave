package models

import (
	"time"
)

// MusicLike model untuk tracking unique likes
type MusicLike struct {
	ID        uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    uint      `json:"user_id" gorm:"not null;index:idx_user_music,unique"`
	MusicID   uint      `json:"music_id" gorm:"not null;index:idx_user_music,unique"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName mengembalikan nama tabel
func (MusicLike) TableName() string {
	return "music_likes"
}
