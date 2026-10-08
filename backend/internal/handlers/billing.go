package handlers

import (
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"sello/internal/database"
	"sello/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ListPlans - GET /api/plans (Public/Auth)
func ListPlans(c *gin.Context) {
    var plans []models.Plan
    if err := database.GetDB().Where("is_active = ?", true).Order("price_monthly ASC").Find(&plans).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans: " + err.Error()})
        return
    }
    c.JSON(http.StatusOK, plans)
}

// GetBankInfo - GET /api/billing/bank
func GetBankInfo(c *gin.Context) {
    var setting models.SystemSetting
    // We assume 'billing_bank_details' key
    if err := database.GetDB().First(&setting, "key = ?", "billing_bank_details").Error; err != nil {
        c.JSON(http.StatusOK, gin.H{"bank_details": ""}) // Return empty if not set
        return
    }
    c.JSON(http.StatusOK, gin.H{"bank_details": setting.Value})
}

// VerifyCoupon - POST /api/billing/coupon/verify
func VerifyCoupon(c *gin.Context) {
    var input struct {
        Code string `json:"code" binding:"required"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var coupon models.Coupon
    // Check platform coupons (store_id IS NULL)
    // Note: We use Unscoped to find even if soft deleted to check validity properly
    if err := database.GetDB().Where("(store_id IS NULL) AND code = ?", input.Code).First(&coupon).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Invalid coupon code: " + err.Error()})
        return
    }

    // Check expiry
    if coupon.ExpiresAt != nil && time.Now().After(*coupon.ExpiresAt) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon expired"})
        return
    }

    // Check usage limit
    if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon usage limit reached"})
        return
    }

    c.JSON(http.StatusOK, coupon)
}

// CreateSubscription - POST /api/billing/subscribe
func CreateSubscription(c *gin.Context) {
	userIdVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
    // Handle user_id type (can be string or uuid.UUID depending on middleware)
    var userID uuid.UUID
    var err error

    switch v := userIdVal.(type) {
    case string:
        userID, err = uuid.Parse(v)
    case uuid.UUID:
        userID = v
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
        return
    }

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
        return
    }

	var input struct {
		PlanID     uint   `json:"plan_id" binding:"required"`
		Duration   int    `json:"duration" binding:"required"` // 1 or 12
        CouponCode string `json:"coupon_code"`
        PaymentMethod string `json:"payment_method"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var plan models.Plan
	if err := database.GetDB().First(&plan, input.PlanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	amount := plan.PriceMonthly * float64(input.Duration)
	if input.Duration >= 12 {
		amount = plan.PriceYearly
		if amount == 0 {
			amount = plan.PriceMonthly * 12
		}
	}

    // Apply Coupon
    var couponID *uint
    if input.CouponCode != "" {
        var coupon models.Coupon
        if err := database.GetDB().Where("store_id IS NULL AND code = ?", input.CouponCode).First(&coupon).Error; err == nil {
             // Validate Expiry & Limit again
             valid := true
             if coupon.ExpiresAt != nil && time.Now().After(*coupon.ExpiresAt) { valid = false }
             if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit { valid = false }
             
             if valid {
                 // Check Min Purchase
                 if coupon.MinPurchase > 0 && amount < coupon.MinPurchase {
                     // Coupon ignore silently
                 } else {
                     discount := 0.0
                     if coupon.Type == "percentage" {
                         discount = amount * (coupon.Value / 100)
                         if coupon.MaxDiscount > 0 && discount > coupon.MaxDiscount {
                             discount = coupon.MaxDiscount
                         }
                     } else {
                         discount = coupon.Value
                     }
                     
                     amount -= discount
                     if amount < 0 { amount = 0 }
                     couponID = &coupon.ID
                     
                     // Increment usage
                     database.GetDB().Model(&coupon).Update("used_count", gorm.Expr("used_count + 1"))
                 }
             }
        }
    }

	sub := models.SubscriptionTransaction{
		UserID:        userID,
		PlanID:        input.PlanID,
		Amount:        amount,
		DurationMonth: input.Duration,
        CouponID:      couponID,
		Status:        models.SubscriptionPending,
        PaymentMethod: input.PaymentMethod,
	}

	if err := database.GetDB().Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}

// UploadSubscriptionProof - POST /api/billing/:id/proof
func UploadSubscriptionProof(c *gin.Context) {
	id := c.Param("id")
	userIdVal, _ := c.Get("user_id")
	var userIDStr string
    switch v := userIdVal.(type) {
    case string:
        userIDStr = v
    case uuid.UUID:
        userIDStr = v.String()
    }

	var sub models.SubscriptionTransaction
	if err := database.GetDB().First(&sub, "id = ? AND user_id = ?", id, userIDStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	file, err := c.FormFile("proof")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	filename := fmt.Sprintf("proof_%s%s", sub.ID, filepath.Ext(file.Filename))
	path := filepath.Join("web", "uploads", filename)

	if err := c.SaveUploadedFile(file, path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	sub.PaymentProof = "/uploads/" + filename
	sub.Status = models.SubscriptionPaid
	database.GetDB().Save(&sub)

	c.JSON(http.StatusOK, sub)
}

// AdminListSubscriptions - GET /console/api/subscriptions
func AdminListSubscriptions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    status := c.Query("status")

	offset := (page - 1) * limit
	var subs []models.SubscriptionTransaction
	var total int64

	db := database.GetDB().Model(&models.SubscriptionTransaction{})
    
    if status != "" {
        db = db.Where("status = ?", status)
    }
    
	db.Count(&total)

	db.Preload("User").Preload("Plan").Order("created_at DESC").Limit(limit).Offset(offset).Find(&subs)

	c.JSON(http.StatusOK, gin.H{
		"data": subs,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// AdminApproveSubscription - PUT /console/api/subscriptions/:id/approve
func AdminApproveSubscription(c *gin.Context) {
	id := c.Param("id")
	var sub models.SubscriptionTransaction
	if err := database.GetDB().Preload("Plan").First(&sub, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if sub.Status == models.SubscriptionActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already active"})
		return
	}

	// Update Transaction
	sub.Status = models.SubscriptionActive
	database.GetDB().Save(&sub)

	// Update User Plan
	var user models.User
	if err := database.GetDB().First(&user, sub.UserID).Error; err == nil {
		user.PlanID = &sub.PlanID

		now := time.Now()
		expiry := now.AddDate(0, sub.DurationMonth, 0)
		user.SubscriptionEndsAt = &expiry

		database.GetDB().Save(&user)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription approved", "user_expiry": user.SubscriptionEndsAt})
}

// AdminCreateManualSubscription - POST /console/api/subscriptions/manual
func AdminCreateManualSubscription(c *gin.Context) {
	var input struct {
		Email         string  `json:"email" binding:"required"`
		PlanID        uint    `json:"plan_id" binding:"required"`
		Duration      int     `json:"duration"` // Default 1
		Amount        float64 `json:"amount"`   // Optional Custom Amount
		Status        string  `json:"status"`   // active, pending, paid
		PaymentMethod string  `json:"payment_method"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Find User by Email
	var user models.User
	if err := database.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
		return
	}

	// 2. Find Plan
	var plan models.Plan
	if err := database.GetDB().First(&plan, input.PlanID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	// 3. Defaults
	if input.Duration < 1 {
		input.Duration = 1
	}
	if input.Status == "" {
		input.Status = string(models.SubscriptionPending)
	}

	// Calculate Amount if not provided
	finalAmount := input.Amount
	if finalAmount <= 0 {
		if input.Duration >= 12 {
			finalAmount = plan.PriceYearly
			if finalAmount == 0 {
				finalAmount = plan.PriceMonthly * 12
			}
		} else {
			finalAmount = plan.PriceMonthly * float64(input.Duration)
		}
	}

	// 4. Create Transaction
	sub := models.SubscriptionTransaction{
		UserID:        user.ID,
		PlanID:        input.PlanID,
		Amount:        finalAmount,
		DurationMonth: input.Duration,
		Status:        models.SubscriptionStatus(input.Status),
		PaymentMethod: input.PaymentMethod,
		// No Coupon for manual
	}

	if err := database.GetDB().Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create subscription"})
		return
	}

	// 5. If Active/Paid, Auto-Approve (Update User Plan)
	if sub.Status == models.SubscriptionActive {
		user.PlanID = &sub.PlanID
		now := time.Now()
		expiry := now.AddDate(0, sub.DurationMonth, 0)
		
		// If user already has active plan and we want to extend? 
		// For simplicity, manual override resets expiry from now.
		// Detailed logic for extension:
		// if user.SubscriptionEndsAt != nil && user.SubscriptionEndsAt.After(now) {
		//    expiry = user.SubscriptionEndsAt.AddDate(0, sub.DurationMonth, 0)
		// }

		user.SubscriptionEndsAt = &expiry
		database.GetDB().Save(&user)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Subscription created successfully",
		"data":    sub,
	})
}
