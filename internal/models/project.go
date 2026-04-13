package models

import (
	"time"

	"gorm.io/gorm"
)

type Folder struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID" json:"-"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Project struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	UserID       uint           `gorm:"not null;index" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"-"`
	FolderID     *uint          `gorm:"index" json:"folder_id"` // Optional mapping
	Folder       Folder         `gorm:"foreignKey:FolderID" json:"-"`
	Title        string         `gorm:"size:255;not null" json:"title"`
	Description  string         `gorm:"type:text" json:"description"`
	IsFavorite   bool           `gorm:"default:false" json:"is_favorite"`
	LastOpenedAt *time.Time     `json:"last_opened_at"`
	ShareToken   *string        `gorm:"size:255" json:"share_token"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
