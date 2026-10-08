package models

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken stores tokens for password reset
type PasswordResetToken struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;index" json:"user_id"`
	Token     string     `gorm:"type:varchar(255);uniqueIndex" json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}
