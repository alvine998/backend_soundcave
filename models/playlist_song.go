package models

import (
	"time"

	"gorm.io/gorm"
)

// PlaylistSong model untuk relasi antara playlist dan music
type PlaylistSong struct {
	ID         uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	PlaylistID uint           `json:"playlist_id" gorm:"not null;index"`
	MusicID    uint           `json:"music_id" gorm:"not null;index"`
	Position   int            `json:"position" gorm:"not null;default:0"` // Urutan lagu dalam playlist
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at" gorm:"index"`

	// Relations (tanpa foreign key constraint untuk menghindari error migration)
	Playlist Playlist `json:"playlist,omitempty" gorm:"foreignKey:PlaylistID;constraint:-"`
	Music    Music    `json:"music,omitempty" gorm:"foreignKey:MusicID;constraint:-"`
}

// TableName mengembalikan nama tabel
func (PlaylistSong) TableName() string {
	return "playlist_songs"
}
