package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a platform user (synced from Supabase auth)
// This is a read-only reference for admin queries
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Email           string         `gorm:"type:text" json:"email"`
	Name            string         `gorm:"type:text" json:"name"`
	AvatarURL       string         `gorm:"type:text" json:"avatar_url"` // From Google/Supabase
	Password        string         `gorm:"type:text" json:"-"`          // Local Auth (Bcrypt hash)
	Whatsapp        *string        `gorm:"type:text;uniqueIndex" json:"whatsapp"`
	IsPhoneVerified bool           `gorm:"default:false" json:"is_phone_verified"`
	IsEmailVerified bool           `gorm:"default:false" json:"is_email_verified"`
	IsSuspended     bool           `gorm:"default:false" json:"is_suspended"`
	
	// Subscription
	PlanID             *uint      `json:"plan_id"`
	SubscriptionEndsAt *time.Time `json:"subscription_ends_at"`
	DowngradePending   bool       `gorm:"default:false" json:"downgrade_pending"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Stores []Store `gorm:"foreignKey:UserID" json:"stores,omitempty"`
	Plan   *Plan   `gorm:"foreignKey:PlanID" json:"plan,omitempty"`
}

func (User) TableName() string {
	return "users"
}
