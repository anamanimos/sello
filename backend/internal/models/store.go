package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Store struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null" json:"user_id"`
	Slug        string         `gorm:"type:text;uniqueIndex;not null" json:"slug"`
	Name        string         `gorm:"type:text;not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Whatsapp    string         `gorm:"type:text" json:"whatsapp"`
	LogoURL     string         `gorm:"type:text" json:"logo_url"`
	FaviconURL  string         `gorm:"type:text" json:"favicon_url"`
	
	// Custom Domain Support
	CustomDomain   *string `gorm:"type:text;uniqueIndex" json:"custom_domain"`
	DomainVerified bool   `gorm:"default:false" json:"domain_verified"`
	VerifyToken    string `gorm:"type:text" json:"verify_token,omitempty"`
	
	// Admin Management
	IsVerified   bool   `gorm:"default:false" json:"is_verified"`
	IsSuspended  bool   `gorm:"default:false" json:"is_suspended"`
	SuspensionReason string `gorm:"type:text" json:"suspension_reason"`
	HideBranding bool   `gorm:"default:false" json:"hide_branding"`
	
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Computed (read-only, populated by SQL select)
	ProductCount int64 `gorm:"column:product_count;->;<-:false" json:"product_count"`

	// Relations
	Products  []Product  `gorm:"foreignKey:StoreID" json:"products,omitempty"`
	Templates []Template `gorm:"foreignKey:StoreID" json:"templates,omitempty"`
}

func (Store) TableName() string {
	return "stores"
}

type CreateStoreInput struct {
	Slug        string `json:"slug" binding:"required,min=3,max=50"`
	Name        string `json:"name" binding:"required,min=1,max=100"`
	Description string `json:"description"`
	Whatsapp    string `json:"whatsapp"`
}

type UpdateStoreInput struct {
	Slug         string `json:"slug" binding:"omitempty,min=3,max=50"`
	Name         string `json:"name" binding:"omitempty,min=1,max=100"`
	Description  string `json:"description"`
	Whatsapp     string `json:"whatsapp"`
	LogoURL      string `json:"logo_url"`
	FaviconURL   string `json:"favicon_url"`
	HideBranding *bool  `json:"hide_branding"` // Pointer to allow false updates
}
