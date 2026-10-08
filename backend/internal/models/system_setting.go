package models

import "time"

// SystemSetting stores global configuration (SMTP, Whatsapp, etc)
type SystemSetting struct {
	Key         string    `gorm:"primaryKey" json:"key"`         // e.g. "smtp_host", "wa_url"
	Value       string    `gorm:"type:text" json:"value"`        // e.g. "smtp.gmail.com"
	Type        string    `gorm:"type:varchar(20)" json:"type"`  // string, int, bool, json
	Description string    `gorm:"type:text" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (SystemSetting) TableName() string {
	return "system_settings"
}
