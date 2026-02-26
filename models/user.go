package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Role enum untuk user
type Role string

const (
	RoleUser        Role = "user"
	RoleAdmin       Role = "admin"
	RolePremium     Role = "premium"
	RoleIndependent Role = "independent"
	RoleLabel       Role = "label"
)

// JSONStringArray type untuk menyimpan JSON array of strings
type JSONStringArray []string

// Value implements driver.Valuer
func (j JSONStringArray) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	bytes, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// Scan implements sql.Scanner
func (j *JSONStringArray) Scan(value interface{}) error {
	if value == nil {
		*j = JSONStringArray{}
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		return fmt.Errorf("cannot scan type %T into JSONStringArray", value)
	}
	return json.Unmarshal(bytes, j)
}

// User model sesuai struktur tabel
type User struct {
	ID            uint            `json:"id" gorm:"primaryKey;autoIncrement"`
	FullName      string          `json:"full_name" gorm:"column:full_name;size:255;not null"`
	Email         string          `json:"email" gorm:"column:email;size:255;not null;uniqueIndex"`
	Password      *string         `json:"-" gorm:"size:255"` // Hidden dari JSON response, nullable untuk Google Auth
	Phone         *string         `json:"phone" gorm:"size:20"`
	Location      *string         `json:"location" gorm:"size:255"`
	Bio           *string         `json:"bio" gorm:"type:text"`
	ProfileImage  *string         `json:"profile_image" gorm:"column:profile_image;size:255"`
	Role          Role            `json:"role" gorm:"type:enum('user','admin','premium','independent','label');default:'user'"`
	Followers     JSONStringArray `json:"followers" gorm:"type:json"`
	TotalFollower int             `json:"total_follower" gorm:"default:0"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	DeletedAt     gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
}

// TableName mengembalikan nama tabel
func (User) TableName() string {
	return "users"
}
