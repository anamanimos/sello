package models

import (
	"time"

	"github.com/google/uuid"
)

type AnalyticsLog struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StoreID   uuid.UUID `gorm:"type:uuid;not null;index" json:"store_id"`
	ProductID *uuid.UUID `gorm:"type:uuid;index" json:"product_id"` // Nullable if page view
	EventType string    `gorm:"type:text;not null" json:"event_type"` // 'view_store', 'view_product', 'click_whatsapp'
	PageURL   string    `gorm:"type:text" json:"page_url"`
	IPAddress string    `gorm:"type:text" json:"ip_address"`
	UserAgent string    `gorm:"type:text" json:"user_agent"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (AnalyticsLog) TableName() string {
	return "analytics_logs"
}

type TrackEventInput struct {
	EventType string `json:"event_type" binding:"required"`
	ProductID string `json:"product_id"` // Optional
	StoreSlug string `json:"store_slug"` // Optional, used if StoreID not provided contextually
}
