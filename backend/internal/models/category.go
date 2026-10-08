package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Category struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StoreID   uuid.UUID      `gorm:"type:uuid;index;not null" json:"store_id"`
	Store     Store          `gorm:"foreignKey:StoreID" json:"-"`
	Name      string         `gorm:"type:text;not null" json:"name"`
	Slug      string         `gorm:"type:text;not null" json:"slug"` // Unique per store constraint handled in logic or compound unique index
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

type CreateCategoryInput struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}

type UpdateCategoryInput struct {
	Name string `json:"name" binding:"required,min=1,max=100"`
}
