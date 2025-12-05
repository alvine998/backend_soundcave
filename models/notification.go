package models

import (
	"time"

	"gorm.io/gorm"
)

// NotificationType enum untuk tipe notification
type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
)

// Notification model sesuai struktur tabel
type Notification struct {
	ID        uint             `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID    int              `json:"user_id" gorm:"not null;index"`
	Title     string           `json:"title" gorm:"size:255;not null"`
	Message   string           `json:"message" gorm:"type:text;not null"`
	Date      time.Time        `json:"date" gorm:"type:date;not null"`
	IsRead    *bool            `json:"is_read" gorm:"type:tinyint(1);default:0"`
	Type      NotificationType `json:"type" gorm:"type:enum('info','success','warning','error');default:'info'"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (Notification) TableName() string {
	return "notifications"
}

