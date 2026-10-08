package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"sello/internal/database"
	"sello/internal/models"
	"sello/internal/services"
)

// GetSettings - GET /console/api/settings/gateway
func GetSettings(c *gin.Context) {
	var settings []models.SystemSetting
	// Fetch relevant settings
	keys := []string{
		// Gateway settings
		"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from",
		"wa_url", "wa_username", "wa_password",
		"verification_default", "password_reset_method",
		// User limits
		"max_users", "max_stores_per_user", "max_products_per_store",
		"allow_registration", "billing_bank_details",
        // Beta Mode
        "beta_mode_enabled", "beta_default_plan_id",
	}
	
	result := database.GetDB().Where("key IN ?", keys).Find(&settings)
	
	// Debug logging
	log.Printf("[GetSettings] Query result: %d rows, Error: %v", result.RowsAffected, result.Error)
	log.Printf("[GetSettings] Settings found: %+v", settings)

	config := make(map[string]string)
	for _, s := range settings {
		config[s.Key] = s.Value
	}

	log.Printf("[GetSettings] Returning config: %+v", config)
	c.JSON(http.StatusOK, config)
}

// UpdateSettings - POST /console/api/settings/gateway
func UpdateSettings(c *gin.Context) {
	var input map[string]string
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for key, value := range input {
		// Use proper upsert with OnConflict
		setting := models.SystemSetting{
			Key:   key,
			Value: value,
			Type:  "string",
		}
		
		// First try to find existing, if not create
		result := database.GetDB().Where("key = ?", key).First(&models.SystemSetting{})
		if result.RowsAffected == 0 {
			// Create new
			database.GetDB().Create(&setting)
		} else {
			// Update existing
			database.GetDB().Model(&models.SystemSetting{}).Where("key = ?", key).Updates(map[string]interface{}{
				"value": value,
				"type":  "string",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// TestGateway - POST /console/api/settings/test
func TestGateway(c *gin.Context) {
	var input struct {
		Type     string `json:"type" binding:"required"` // email or whatsapp
		Target   string `json:"target" binding:"required"`
		Template string `json:"template"` // otp, password_reset, magic_link, welcome (optional, defaults to otp)
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var err error
	if input.Type == "whatsapp" {
		// Send generic test message
		err = services.SendWhatsappOTP(input.Target, "TEST12 (System Test)")
	} else {
		// Email with template selection
		template := input.Template
		if template == "" {
			template = "otp"
		}

		switch template {
		case "password_reset":
			err = services.SendPasswordResetEmail(input.Target, "https://sello.id/reset?token=TEST123")
		case "magic_link":
			err = services.SendMagicLinkEmail(input.Target, "https://sello.id/login?token=TEST123")
		case "welcome":
			err = services.SendWelcomeEmail(input.Target, "User Test")
		default: // otp
			err = services.SendEmailOTP(input.Target, "123456")
		}
	}

	if err != nil {
		log.Printf("[TestGateway] Error sending %s to %s: %v", input.Type, input.Target, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test message sent successfully"})
}

// PreviewEmailTemplate - GET /console/api/settings/email-preview?template=otp
func PreviewEmailTemplate(c *gin.Context) {
	template := c.Query("template")
	if template == "" {
		template = "otp"
	}

	var html string
	switch template {
	case "password_reset":
		html = services.RenderPasswordResetEmail("https://sello.id/reset?token=PREVIEW123")
	case "magic_link":
		html = services.RenderMagicLinkEmail("https://sello.id/login?token=PREVIEW123")
	case "welcome":
		html = services.RenderWelcomeEmail("User Preview")
	default: // otp
		html = services.RenderOTPEmail("123456")
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}
