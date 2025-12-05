package models

import (
	"time"

	"gorm.io/gorm"
)

// Image model untuk menyimpan informasi gambar yang diupload
type Image struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	FileName    string         `json:"file_name" gorm:"size:255;not null"`
	FileURL     string         `json:"file_url" gorm:"size:500;not null"`
	FileSize    int64          `json:"file_size"`
	ContentType string         `json:"content_type" gorm:"size:100"`
	BucketPath  string         `json:"bucket_path" gorm:"size:500"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

