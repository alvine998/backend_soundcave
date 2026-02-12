package models

import (
	"time"

	"gorm.io/gorm"
)

// AlbumType enum untuk tipe album
type AlbumType string

const (
	AlbumTypeSingle      AlbumType = "single"
	AlbumTypeEP          AlbumType = "EP"
	AlbumTypeAlbum       AlbumType = "album"
	AlbumTypeCompilation AlbumType = "compilation"
)

// Album model sesuai struktur tabel
type Album struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Title       string         `json:"title" gorm:"size:255;not null"`
	ArtistID    int            `json:"artist_id" gorm:"not null;index"`
	Artist      string         `json:"artist" gorm:"size:255;not null"`
	ReleaseDate *time.Time     `json:"release_date" gorm:"type:date"`
	AlbumType   AlbumType      `json:"album_type" gorm:"type:enum('single','EP','album','compilation');not null"`
	Genre       string         `json:"genre" gorm:"size:100"`
	TotalTracks int            `json:"total_tracks" gorm:"not null;default:0"`
	RecordLabel *string        `json:"record_label" gorm:"size:255"`
	Image       *string        `json:"image" gorm:"size:255"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Album) TableName() string {
	return "albums"
}
