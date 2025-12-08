package models

import (
	"time"

	"gorm.io/gorm"
)

// Role enum untuk user
type Role string

const (
	RoleUser    Role = "user"
	RoleAdmin   Role = "admin"
	RolePremium Role = "premium"
)

// User model sesuai struktur tabel
type User struct {
	ID           uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	FullName     string         `json:"full_name" gorm:"size:255;not null"`
	Email        string         `json:"email" gorm:"column:email;size:255;not null;uniqueIndex"`
	Password     *string        `json:"-" gorm:"size:255"` // Hidden dari JSON response, nullable untuk Google Auth
	Phone        *string        `json:"phone" gorm:"size:20"`
	Location     *string        `json:"location" gorm:"size:255"`
	Bio          *string        `json:"bio" gorm:"type:text"`
	ProfileImage *string        `json:"profile_image" gorm:"column:profile_image;size:255"`
	Role         Role           `json:"role" gorm:"type:enum('user','admin','premium');default:'user'"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (User) TableName() string {
	return "users"
}
