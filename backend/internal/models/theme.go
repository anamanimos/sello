package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Theme (Disk Based) - Struct untuk response API, tidak disimpan di DB
type Theme struct {
	ID             string                 `json:"id"` // Folder name as ID/Slug
	Slug           string                 `json:"slug"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	ThumbnailURL   string                 `json:"thumbnail_url"`
	Version        string                 `json:"version"`
	DirectoryPath  string                 `json:"directory_path"`
	IsActive       bool                   `json:"is_active,omitempty"` // Contextual (for specific store)
	IsPremium      bool                   `json:"is_premium"`
	IsHidden       bool                   `json:"is_hidden"` // Exclusive/Hidden theme
	SortOrder      int                    `json:"sort_order"`
	AllowedUserIDs []string               `json:"allowed_user_ids,omitempty"` // User IDs that can use this theme
	DefaultConfig  map[string]interface{} `json:"default_config"`
}

// StoreTheme - Tema yang dipilih dan dikustomisasi oleh user
// Hanya menyimpan referensi nama folder tema (ThemeName)
type StoreTheme struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoreID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"store_id"`
	ThemeName    string         `gorm:"size:100;not null;index" json:"theme_name"` // Nama folder tema (e.g. "starter")
	CustomConfig datatypes.JSON `gorm:"type:jsonb" json:"custom_config"`
	IsActive     bool           `gorm:"default:false" json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	// Store Relation
	Store *Store `gorm:"foreignKey:StoreID" json:"-"`
	
	// Virtual field for API response (populated manually from disk)
	Theme *Theme `gorm:"-" json:"theme,omitempty"`
}

// StoreThemeHistory - Riwayat konfigurasi
type StoreThemeHistory struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	StoreThemeID uuid.UUID      `gorm:"type:uuid;not null;index" json:"store_theme_id"`
	CustomConfig datatypes.JSON `gorm:"type:jsonb" json:"custom_config"`
	Label        string         `gorm:"size:100" json:"label"`
	CreatedAt    time.Time      `json:"created_at"`

	StoreTheme *StoreTheme `gorm:"foreignKey:StoreThemeID" json:"store_theme,omitempty"`
}

// Helper structs for config (Disk Schema)
type ThemeConfigSchema struct {
	Settings []ThemeSettingGroup `json:"settings"`
}

type ThemeSettingGroup struct {
	ID       string         `json:"id"`
	Title    string         `json:"title"`
	Sections []ThemeSection `json:"sections"`
}

type ThemeSection struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Settings []ThemeSetting `json:"settings"`
}

type ThemeSetting struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Label       string                 `json:"label"`
	Content     string                 `json:"content,omitempty"` // For headers
	Default     interface{}            `json:"default"`
	Options     []ThemeOption          `json:"options,omitempty"`
	Placeholder string                 `json:"placeholder,omitempty"`
	HelpText    string                 `json:"help_text,omitempty"`
	Min         float64                `json:"min,omitempty"`
	Max         float64                `json:"max,omitempty"`
	Step        float64                `json:"step,omitempty"`
	Condition   map[string]interface{} `json:"condition,omitempty"`
	Schema      []ThemeSetting         `json:"schema,omitempty"` // For Repeater types
}

type ThemeOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// ThemeMeta - Metadata from theme.json (Root)
type ThemeMeta struct {
	Name          string                 `json:"name"`
	Slug          string                 `json:"slug"`
	Description   string                 `json:"description"`
	Version       string                 `json:"version"`
	ThumbnailURL  string                 `json:"thumbnail_url"`
	SortOrder     int                    `json:"sort_order"`
	IsPremium     bool                   `json:"is_premium"`
	IsHidden       bool                   `json:"is_hidden"`
	AllowedUserIDs []string               `json:"allowed_user_ids,omitempty"` // User IDs that can use this theme
	Templates      map[string]string      `json:"templates"`       // Map page type to filename
	Settings       []ThemeSetting         `json:"settings"`        // Global settings schema
	DefaultConfig  ThemeConfig            `json:"default_config"`  // Default values
}

// ThemeConfig - Validated config map
type ThemeConfig map[string]interface{}

