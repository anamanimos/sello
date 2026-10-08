package models

import (
	"time"

	"github.com/google/uuid"
)

type OTP struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Code      string    `gorm:"type:varchar(10);not null" json:"code"`
	Type      string    `gorm:"type:varchar(20)" json:"type"` // "email" or "whatsapp"
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `gorm:"default:false" json:"used"`
	
	// Relation
	User User `gorm:"foreignKey:UserID" json:"-"`
}

func (OTP) TableName() string {
	return "otps"
}
