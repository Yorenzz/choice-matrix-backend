package models

import (
	"time"

	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
	UserStatusPending  UserStatus = "pending"
)

type UserPlan string

const (
	UserPlanFree UserPlan = "free"
	UserPlanPro  UserPlan = "pro"
)

type User struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Email            string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash     string         `gorm:"not null" json:"-"`
	Nickname         string         `gorm:"size:255" json:"nickname"`
	AvatarURL        string         `gorm:"size:1024" json:"avatar_url"`
	Status           UserStatus     `gorm:"type:varchar(20);default:'active';index" json:"status"`
	Plan             UserPlan       `gorm:"type:varchar(20);default:'free';index" json:"plan"`
	PlanExpiresAt    *time.Time     `json:"plan_expires_at"`
	EmailVerifiedAt  *time.Time     `json:"email_verified_at"`
	AICredits        int            `gorm:"default:10" json:"ai_credits"`
	AICreditsResetAt *time.Time     `json:"ai_credits_reset_at"`
	LastLoginAt      *time.Time     `json:"last_login_at"`
	LastLoginIP      string         `gorm:"size:64" json:"last_login_ip"`
	CountryCode      string         `gorm:"size:8" json:"country_code"`
	RegionName       string         `gorm:"size:128" json:"region_name"`
	CityName         string         `gorm:"size:128" json:"city_name"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}
