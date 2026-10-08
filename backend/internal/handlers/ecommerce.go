package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sello/internal/database"
	"sello/internal/models"
)

// CreateOrder - POST /api/stores/:store_id/checkout
func CreateOrder(c *gin.Context) {
	storeID := c.Param("store_id")

	var input models.CreateOrderInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Store
	var store models.Store
	if err := database.GetDB().Where("id = ?", storeID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Calculate Total & Validate Items
	var totalAmount float64
	var orderItems []models.OrderItem

	for _, item := range input.Items {
		var product models.Product
		if err := database.GetDB().Where("id = ? AND store_id = ?", item.ProductID, store.ID).First(&product).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Product not found: %s", item.ProductID)})
			return
		}

		// Determine price (check discount)
		price := float64(product.Price)
		if product.DiscountPrice != nil && *product.DiscountPrice > 0 {
			price = float64(*product.DiscountPrice)
		}

		subtotal := price * float64(item.Quantity)
		totalAmount += subtotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID:   product.ID,
			ProductName: product.Name,
			VariantName: item.Variant,
			Price:       price,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	// Apply Coupon
	discountAmount := 0.0
	var couponID *uint

	if input.CouponCode != "" {
		var coupon models.Coupon
		// Check Store-Specific Coupon
		if err := database.GetDB().Where("store_id = ? AND code = ?", store.ID, input.CouponCode).First(&coupon).Error; err == nil {
			// Validate
			now := time.Now()
			if coupon.ExpiresAt != nil && coupon.ExpiresAt.Before(now) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon expired"})
				return
			}
			if coupon.UsageLimit > 0 && coupon.UsedCount >= coupon.UsageLimit {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Coupon usage limit reached"})
				return
			}
			if totalAmount < coupon.MinPurchase {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Minimum purchase for coupon is %.2f", coupon.MinPurchase)})
				return
			}

			// Calculate Discount
			if coupon.Type == models.CouponTypePercentage {
				dic := (coupon.Value / 100) * totalAmount
				if coupon.MaxDiscount > 0 && dic > coupon.MaxDiscount {
					dic = coupon.MaxDiscount
				}
				discountAmount = dic
			} else {
				discountAmount = coupon.Value // Fixed amount
			}
            
            // Ensure discount doesn't exceed total
            if discountAmount > totalAmount {
                discountAmount = totalAmount
            }

			couponID = &coupon.ID
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid coupon code"})
			return
		}
	}

	finalAmount := totalAmount - discountAmount

	// Create Order
	order := models.Order{
		StoreID:        store.ID,
		CustomerName:   input.CustomerName,
		CustomerPhone:  input.CustomerPhone,
		CustomerEmail:  input.CustomerEmail,
		TotalAmount:    totalAmount,
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Status:         models.OrderStatusPending, // Pending Manual Payment
		PaymentMethod:  input.PaymentMethod,
		CouponCode:     input.CouponCode,
		CouponID:       couponID,
		Items:          orderItems,
	}

	// Transaction to ensure atomic Coupon Usage increment
	tx := database.GetDB().Begin()
	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Update Coupon Usage
	if couponID != nil {
		if err := tx.Model(&models.Coupon{}).Where("id = ?", *couponID).UpdateColumn("used_count", gorm.Expr("used_count + ?", 1)).Error; err != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply coupon"})
            return
        }
	}
	
	tx.Commit()

	c.JSON(http.StatusCreated, order)
}

// UploadPaymentProof - POST /api/orders/:order_id/pay
func UploadPaymentProof(c *gin.Context) {
	orderID := c.Param("order_id")

	var order models.Order
	if err := database.GetDB().First(&order, "id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check status
	if order.Status != models.OrderStatusPending {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is not pending payment"})
		return
	}

	file, header, err := c.Request.FormFile("proof")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No proof image provided"})
		return
	}
	defer file.Close()
    
    // Validate File
    ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type (jpg, png common only)"})
        return
    }

	// Initialize Cloudinary
	if cld == nil {
		if err := InitCloudinary(); err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to init upload service"})
            return
        }
	}
    
    // Double check
    if cld == nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload service unavailable"})
         return
    }

	ctx := context.Background()
	// Public ID: sello/proofs/{order_id}
	publicID := fmt.Sprintf("sello/proofs/%s_%d", order.ID, time.Now().Unix())

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID: publicID,
		Folder:   "sello/proofs",
        Transformation: "c_limit,w_1000,h_2000,q_auto", // Optimize proof image
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upload failed"})
		return
	}

	// Update Order
	order.PaymentProof = uploadResult.SecureURL
	order.Status = models.OrderStatusWaitingConfirmation
	database.GetDB().Save(&order)

	c.JSON(http.StatusOK, gin.H{"message": "Payment proof uploaded", "order": order})
}

// GetOrder - GET /api/orders/:order_id
func GetOrder(c *gin.Context) {
	orderID := c.Param("order_id")
	var order models.Order
	if err := database.GetDB().Preload("Items").First(&order, "id = ?", orderID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// GetStoreOrders - GET /api/stores/:store_id/orders
func GetStoreOrders(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify Store Ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var orders []models.Order
	database.GetDB().Where("store_id = ?", store.ID).Preload("Items").Order("created_at desc").Find(&orders)

	c.JSON(http.StatusOK, orders)
}

// UpdateOrderStatus - PUT /api/stores/:store_id/orders/:order_id/status
func UpdateOrderStatus(c *gin.Context) {
	storeID := c.Param("store_id")
	orderID := c.Param("order_id")
	userID, _ := c.Get("user_id")

	// Verify Store Ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var input struct {
		Status models.OrderStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := database.GetDB().Where("id = ? AND store_id = ?", orderID, store.ID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	order.Status = input.Status
	database.GetDB().Save(&order)

	c.JSON(http.StatusOK, order)
}
