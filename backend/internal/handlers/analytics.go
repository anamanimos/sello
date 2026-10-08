package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// Helper to track event (called internally)
func TrackEvent(c *gin.Context, storeID uuid.UUID, eventType string, productID *uuid.UUID) {
	// Simple async tracking to not block response
	go func() {
		log := models.AnalyticsLog{
			StoreID:   storeID,
			EventType: eventType,
			ProductID: productID,
			PageURL:   c.Request.URL.String(),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		}
		database.GetDB().Create(&log)
	}()
}

// TrackClick - POST /api/analytics/track
// Public endpoint for frontend to track clicks (e.g. WhatsApp)
func TrackClick(c *gin.Context) {
	var input models.TrackEventInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var store models.Store
	// Try to find store by slug (passed from frontend)
	if err := database.GetDB().Where("slug = ?", input.StoreSlug).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var productID *uuid.UUID
	if input.ProductID != "" {
		pid, err := uuid.Parse(input.ProductID)
		if err == nil {
			productID = &pid
		}
	}

	TrackEvent(c, store.ID, input.EventType, productID)
	c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}

// GetDashboardStats - GET /api/stores/:store_id/stats
func GetDashboardStats(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// 1. Total Visits (view_store + view_product)
	var visitCount int64
	database.GetDB().Model(&models.AnalyticsLog{}).
		Where("store_id = ? AND event_type IN ('view_store', 'view_product')", store.ID).
		Count(&visitCount)

	// 2. WhatsApp Clicks
	var waCount int64
	database.GetDB().Model(&models.AnalyticsLog{}).
		Where("store_id = ? AND event_type = 'click_whatsapp'", store.ID).
		Count(&waCount)

	// 3. Product Count
	var productCount int64
	database.GetDB().Model(&models.Product{}).
		Where("store_id = ?", store.ID).
		Count(&productCount)
	
	// 4. Storage Usage
	var totalSize int64
	database.GetDB().Model(&models.Product{}).
		Where("store_id = ?", store.ID).
		Select("COALESCE(SUM(image_size), 0)").
		Scan(&totalSize)
	
	// Add Product Images size
	var imagesSize int64
	database.GetDB().Model(&models.ProductImage{}).
		Joins("JOIN products ON products.id = product_images.product_id").
		Where("products.store_id = ?", store.ID).
		Select("COALESCE(SUM(product_images.image_size), 0)").
		Scan(&imagesSize)
	
	totalSize += imagesSize

	// Limit 30 days history for chart (optional, maybe later)
	
	c.JSON(http.StatusOK, gin.H{
		"visits":       visitCount,
		"clicks":       waCount,
		"products":     productCount,
		"storage_used": totalSize,
		"storage_max":  50 * 1024 * 1024,
	})
}
