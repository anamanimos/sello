package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sello/internal/database"
	"sello/internal/models"
)

// SetCustomDomainInput - Input for setting custom domain
type SetCustomDomainInput struct {
	Domain string `json:"domain" binding:"required"`
}

// SetCustomDomain - PUT /api/stores/:store_id/domain
// Sets a custom domain for the store and generates verification token
func SetCustomDomain(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var input SetCustomDomainInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean domain input
	domain := strings.ToLower(strings.TrimSpace(input.Domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")

	// Validate domain format
	if !isValidDomain(domain) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid domain format"})
		return
	}

	// Check if domain is already used by another store
	var existingStore models.Store
	if err := database.GetDB().Where("custom_domain = ? AND id != ?", domain, store.ID).First(&existingStore).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Domain already in use by another store"})
		return
	}

	// Generate verification token
	token := generateVerifyToken()

	// Update store
	store.CustomDomain = &domain
	store.DomainVerified = false
	store.VerifyToken = token

	if err := database.GetDB().Save(&store).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":       domain,
		"verified":     false,
		"verify_token": token,
		"instructions": gin.H{
			"step1": "Add a CNAME record pointing your domain to stores.sello.id",
			"step2": "Add a TXT record with name '_sello' and value '" + token + "'",
			"step3": "Wait for DNS propagation (may take up to 48 hours)",
			"step4": "Click 'Verify Domain' to complete setup",
		},
	})
}

// VerifyCustomDomain - POST /api/stores/:store_id/domain/verify
// Verifies DNS configuration for custom domain
func VerifyCustomDomain(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	if store.CustomDomain == nil || *store.CustomDomain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No custom domain set"})
		return
	}

	if store.DomainVerified {
		c.JSON(http.StatusOK, gin.H{
			"verified": true,
			"message":  "Domain already verified",
		})
		return
	}

	// Check DNS TXT record
	txtRecords, err := net.LookupTXT("_sello." + *store.CustomDomain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"verified": false,
			"error":    "Could not find TXT record. Please add TXT record for _sello." + *store.CustomDomain,
		})
		return
	}

	// Check if token matches
	tokenFound := false
	for _, record := range txtRecords {
		if strings.TrimSpace(record) == store.VerifyToken {
			tokenFound = true
			break
		}
	}

	if !tokenFound {
		c.JSON(http.StatusBadRequest, gin.H{
			"verified":       false,
			"error":          "TXT record value does not match verification token",
			"expected_value": store.VerifyToken,
		})
		return
	}

	// Optionally check CNAME (but not required for verification)
	// CNAME check can fail if user uses A record instead

	// Mark as verified
	store.DomainVerified = true
	database.GetDB().Save(&store)

	c.JSON(http.StatusOK, gin.H{
		"verified": true,
		"message":  "Domain verified successfully! Your store is now accessible at https://" + *store.CustomDomain,
	})
}

// RemoveCustomDomain - DELETE /api/stores/:store_id/domain
// Removes custom domain from store
func RemoveCustomDomain(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	store.CustomDomain = nil
	store.DomainVerified = false
	store.VerifyToken = ""

	if err := database.GetDB().Save(&store).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom domain removed"})
}

// GetDomainStatus - GET /api/stores/:store_id/domain
// Gets custom domain status
func GetDomainStatus(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	if store.CustomDomain == nil || *store.CustomDomain == "" {
		c.JSON(http.StatusOK, gin.H{
			"has_domain": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"has_domain":   true,
		"domain":       store.CustomDomain,
		"verified":     store.DomainVerified,
		"verify_token": store.VerifyToken,
		"subdomain":    store.Slug + ".sello.id",
	})
}

// Helper functions

func isValidDomain(domain string) bool {
	// Basic domain validation
	if len(domain) < 4 || len(domain) > 253 {
		return false
	}
	if !strings.Contains(domain, ".") {
		return false
	}
	// Check for invalid characters
	for _, c := range domain {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '.') {
			return false
		}
	}
	// Cannot start or end with hyphen
	if strings.HasPrefix(domain, "-") || strings.HasSuffix(domain, "-") {
		return false
	}
	return true
}

func generateVerifyToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "sello-verify-" + hex.EncodeToString(bytes)
}
