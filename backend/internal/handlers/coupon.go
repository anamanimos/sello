package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// CreateCoupon - POST /api/stores/:store_id/coupons
func CreateCoupon(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify Store
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var input models.CreateCouponInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check duplication
	var existing models.Coupon
	if err := database.GetDB().Where("store_id = ? AND code = ?", store.ID, input.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Coupon code already exists"})
		return
	}

	// Parse Expiry
	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			// Try YYYY-MM-DD
			t2, err2 := time.Parse("2006-01-02", input.ExpiresAt)
			if err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD or RFC3339"})
				return
			}
			// End of day
			t2 = t2.Add(23*time.Hour + 59*time.Minute)
			expiresAt = &t2
		} else {
			expiresAt = &t
		}
	}

	storeUUID := store.ID

	coupon := models.Coupon{
		StoreID:     &storeUUID,
		Code:        input.Code,
		Type:        input.Type,
		Value:       input.Value,
		MinPurchase: input.MinPurchase,
		MaxDiscount: input.MaxDiscount,
		UsageLimit:  input.UsageLimit,
		ExpiresAt:   expiresAt,
	}

	if err := database.GetDB().Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon"})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// GetCoupons - GET /api/stores/:store_id/coupons
func GetCoupons(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify Store
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var coupons []models.Coupon
	database.GetDB().Where("store_id = ?", store.ID).Order("created_at desc").Find(&coupons)

	c.JSON(http.StatusOK, coupons)
}

// DeleteCoupon - DELETE /api/coupons/:id
func DeleteCoupon(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var coupon models.Coupon
	if err := database.GetDB().First(&coupon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Coupon not found"})
		return
	}

	// Verify Access
	if coupon.StoreID != nil {
		var store models.Store
		if err := database.GetDB().Where("id = ?", *coupon.StoreID).First(&store).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
			return
		}
		if store.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
			return
		}
	} else {
		// Platform coupon - only admin (not implemented for user)
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete platform coupon"})
		return
	}

	database.GetDB().Delete(&coupon)
	c.JSON(http.StatusOK, gin.H{"message": "Coupon deleted"})
}
