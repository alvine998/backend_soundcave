package models

import (
	"time"

	"gorm.io/gorm"
)

// Genre model sesuai struktur tabel
type Genre struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string         `json:"name" gorm:"size:100;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text;not null"`
	Color       *string        `json:"color" gorm:"size:7"` // Hex color code (e.g., #FF5733)
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Genre) TableName() string {
	return "genres"
}

