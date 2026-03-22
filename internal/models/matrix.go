package models

import (
	"time"

	"gorm.io/gorm"
)

type Row struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	ProjectID     uint           `gorm:"not null;index" json:"project_id"`
	Project       Project        `gorm:"foreignKey:ProjectID" json:"-"`
	Name          string         `gorm:"size:255;not null" json:"name"`
	CoverImageURL string         `gorm:"size:1024" json:"cover_image_url"`
	SortOrder     int            `gorm:"default:0" json:"sort_order"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

type ColumnType string

const (
	ColumnTypeText    ColumnType = "text"
	ColumnTypeNumeric ColumnType = "numeric"
	ColumnTypeScore   ColumnType = "score" // Smart Scoring Column
	ColumnTypeSelect  ColumnType = "select"
)

type Column struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProjectID uint           `gorm:"not null;index" json:"project_id"`
	Project   Project        `gorm:"foreignKey:ProjectID" json:"-"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Type      ColumnType     `gorm:"type:varchar(20);default:'text'" json:"type"`
	Weight    float64        `gorm:"default:1.0" json:"weight"`
	Config    string         `gorm:"type:jsonb" json:"config"` // Using map[string]interface{} via generic JSONB can be done, but string is simpler to unmarshal as generic JSONB in GORM
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Cell struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ProjectID    uint           `gorm:"not null;index" json:"project_id"`
	RowID        uint           `gorm:"not null;uniqueIndex:idx_row_col" json:"row_id"`
	ColumnID     uint           `gorm:"not null;uniqueIndex:idx_row_col" json:"column_id"`
	Row          Row            `gorm:"foreignKey:RowID" json:"-"`
	Column       Column         `gorm:"foreignKey:ColumnID" json:"-"`
	TextContent  string         `gorm:"type:text" json:"text_content"`
	NumericValue float64        `json:"numeric_value"`
	ScoreValue   float64        `json:"score_value"` // 1-10 rating
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}
