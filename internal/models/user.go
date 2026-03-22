package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Email         string         `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string         `gorm:"not null" json:"-"`
	Nickname      string         `gorm:"size:255" json:"nickname"`
	ProStatus     bool           `gorm:"default:false" json:"pro_status"`
	ProExpiryDate *time.Time     `json:"pro_expiry_date"`
	AICredits     int            `gorm:"default:10" json:"ai_credits"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
