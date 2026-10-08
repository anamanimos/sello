package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StoreID       uuid.UUID      `gorm:"type:uuid;not null;index:idx_store_slug,unique" json:"store_id"`
	Name          string         `gorm:"type:text;not null" json:"name"`
	Slug          string         `gorm:"type:text;index:idx_store_slug,unique" json:"slug"` // Composite unique index with StoreID

	Price         int64          `gorm:"type:bigint;not null;default:0" json:"price"`
	DiscountPrice *int64         `gorm:"type:bigint" json:"discount_price"`
	Description   string         `gorm:"type:text" json:"description"`
	ImageURL      string         `gorm:"type:text" json:"image_url"` // Primary image (backward compat)
	ImageSize     int64          `gorm:"type:bigint;default:0" json:"image_size"`
	ExternalLink  string         `gorm:"type:text;default:''" json:"external_link"` // New feature
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Store      Store          `gorm:"foreignKey:StoreID" json:"-"`
	Images     []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	CategoryID *uuid.UUID     `gorm:"type:uuid;index" json:"category_id"`
	Category   *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// ProductImage - Multiple images per product
type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	ImageURL  string    `gorm:"type:text;not null" json:"image_url"`
	ImageSize int64     `gorm:"type:bigint;default:0" json:"image_size"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (ProductImage) TableName() string {
	return "product_images"
}

type CreateProductInput struct {
	Name          string `json:"name" binding:"required,min=1,max=200"`
	Price         int64  `json:"price" binding:"gte=0"`
	DiscountPrice *int64 `json:"discount_price"`
	Description   string `json:"description"`
	ImageURL      string `json:"image_url"`
	ImageSize     int64  `json:"image_size"`
	ExternalLink  string `json:"external_link"`
	CategoryID    string `json:"category_id"` // Receive as string (UUID)
}

type UpdateProductInput struct {
	Name          string  `json:"name" binding:"omitempty,min=1,max=200"`
	Price         int64   `json:"price" binding:"omitempty,gte=0"`
	DiscountPrice *int64  `json:"discount_price"`
	Description   string  `json:"description"`
	ImageURL      *string `json:"image_url"`
	ImageSize     int64   `json:"image_size"`
	ExternalLink  *string `json:"external_link"`
	CategoryID    *string `json:"category_id"`
}
