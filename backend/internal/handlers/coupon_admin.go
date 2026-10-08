package handlers

import (
	"net/http"
	"time"

	"sello/internal/database"
	"sello/internal/models"

	"github.com/gin-gonic/gin"
)

// AdminListCoupons - GET /console/api/coupons
func AdminListCoupons(c *gin.Context) {
	var coupons []models.Coupon
	// Platform coupons have store_id IS NULL
	if err := database.GetDB().Where("store_id IS NULL").Order("created_at desc").Find(&coupons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch coupons: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, coupons)
}

// AdminCreateCoupon - POST /console/api/coupons
func AdminCreateCoupon(c *gin.Context) {
	var input models.CreateCouponInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check duplication (Platform wide)
	var existing models.Coupon
	if err := database.GetDB().Where("store_id IS NULL AND code = ?", input.Code).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Coupon code already exists"})
		return
	}

	// Parse Expiry
	var expiresAt *time.Time
	if input.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, input.ExpiresAt)
		if err != nil {
			t2, err2 := time.Parse("2006-01-02", input.ExpiresAt)
			if err2 != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
				return
			}
			t2 = t2.Add(23*time.Hour + 59*time.Minute)
			expiresAt = &t2
		} else {
			expiresAt = &t
		}
	}

	coupon := models.Coupon{
		StoreID:     nil, // Platform Coupon
		Code:        input.Code,
		Type:        input.Type,
		Value:       input.Value,
		MinPurchase: input.MinPurchase,
		MaxDiscount: input.MaxDiscount,
		UsageLimit:  input.UsageLimit,
		ExpiresAt:   expiresAt,
	}

	if err := database.GetDB().Create(&coupon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create coupon: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, coupon)
}

// AdminDeleteCoupon - DELETE /console/api/coupons/:id
func AdminDeleteCoupon(c *gin.Context) {
	id := c.Param("id")
	if err := database.GetDB().Where("store_id IS NULL").Delete(&models.Coupon{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete coupon"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Coupon deleted"})
}
