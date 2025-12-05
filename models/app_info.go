package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// JSONB type untuk menyimpan JSON data
type JSONB map[string]interface{}

// Value implements driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, j)
}

// AppInfo model sesuai struktur tabel
type AppInfo struct {
	ID          uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	AppName     string         `json:"app_name" gorm:"size:255;not null"`
	Tagline     string         `json:"tagline" gorm:"size:255;not null"`
	Description string         `json:"description" gorm:"type:text;not null"`
	Version     *string        `json:"version" gorm:"size:50"`
	LaunchDate  *time.Time     `json:"launch_date" gorm:"type:date"`
	Email       string         `json:"email" gorm:"size:255;not null"`
	Phone       string         `json:"phone" gorm:"size:50;not null"`
	Address     string         `json:"address" gorm:"type:text;not null"`
	SocialMedia JSONB          `json:"social_media" gorm:"type:json"`
	AppLinks    JSONB          `json:"app_links" gorm:"type:json"`
	Legal       JSONB          `json:"legal" gorm:"type:json"`
	Features    JSONB          `json:"features" gorm:"type:json"`
	Stats       JSONB          `json:"stats" gorm:"type:json"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (AppInfo) TableName() string {
	return "app_info"
}

