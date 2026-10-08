package models

import (
	"time"

	"gorm.io/gorm"
)

// Plan represents a subscription plan
type Plan struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	
	// Limits
	MaxStores           int `gorm:"default:1" json:"max_stores"`
	MaxProductsPerStore int `gorm:"default:10" json:"max_products_per_store"`
	MaxStorageMB        int `gorm:"default:100" json:"max_storage_mb"`
	
	// Features
	AllowCustomDomain   bool `gorm:"default:false" json:"allow_custom_domain"`
	AllowCustomSlug     bool `gorm:"default:false" json:"allow_custom_slug"`
	AllowExternalLink   bool `gorm:"default:false" json:"allow_external_link"`
	RemoveBranding      bool `gorm:"default:false" json:"remove_branding"`
	AccessPremiumThemes bool `gorm:"default:false" json:"access_premium_themes"`
    IsActive            bool `gorm:"default:true" json:"is_active"` // For grandfathering
	
	// Pricing
	PriceMonthly float64 `gorm:"type:decimal(10,2);default:0" json:"price_monthly"`
	PriceYearly  float64 `gorm:"type:decimal(10,2);default:0" json:"price_yearly"`
	
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	// Relations
	Users []User `gorm:"foreignKey:PlanID" json:"-"`
}

func (Plan) TableName() string {
	return "plans"
}
