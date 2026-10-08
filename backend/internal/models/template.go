package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Template struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	StoreID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"store_id"`
	Version   int            `gorm:"type:int;not null;default:1" json:"version"`
	HTMLURL   string         `gorm:"type:text" json:"html_url"`
	CSSURL    string         `gorm:"type:text" json:"css_url"`
	JSURL     string         `gorm:"type:text" json:"js_url"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations
	Store Store `gorm:"foreignKey:StoreID" json:"store,omitempty"`
}

func (Template) TableName() string {
	return "templates"
}

type CreateTemplateInput struct {
	HTMLURL string `json:"html_url"`
	CSSURL  string `json:"css_url"`
	JSURL   string `json:"js_url"`
}

type UpdateTemplateInput struct {
	HTMLURL string `json:"html_url"`
	CSSURL  string `json:"css_url"`
	JSURL   string `json:"js_url"`
}
