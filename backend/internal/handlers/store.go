package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// GetStoreBySlug - GET /store/:slug
func GetStoreBySlug(c *gin.Context) {
	slug := c.Param("slug")

	var store models.Store
	result := database.GetDB().Preload("Products").Where("slug = ?", slug).First(&store)

	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	c.JSON(http.StatusOK, store)
}

// CreateStore - POST /stores
func CreateStore(c *gin.Context) {
	var input models.CreateStoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user_id from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Check if slug already exists
	var existingStore models.Store
	if database.GetDB().Where("slug = ?", input.Slug).First(&existingStore).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Slug already taken"})
		return
	}

	// CHECK PLAN LIMITS
	var user models.User
	if err := database.GetDB().Preload("Plan").First(&user, userID).Error; err == nil {
		maxStores := 1 // Default for No Plan
		if user.Plan != nil {
			maxStores = user.Plan.MaxStores
		} else {
			// Check system setting fallback
			var sysSetting models.SystemSetting
			if database.GetDB().Where("key = ?", "max_stores_per_user").First(&sysSetting).Error == nil {
				fmt.Sscanf(sysSetting.Value, "%d", &maxStores)
			}
		}
		
		var currentStores int64
		database.GetDB().Model(&models.Store{}).Where("user_id = ?", userID).Count(&currentStores)
		
		if currentStores >= int64(maxStores) {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Batas maksimum toko (%d) tercapai. Upgrade paket Anda.", maxStores)})
			return
		}
	}

	store := models.Store{
		UserID:   uuid.MustParse(userID.(string)),
		Slug:     input.Slug,
		Name:     input.Name,
		Whatsapp: input.Whatsapp,
	}

	if err := database.GetDB().Create(&store).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create store: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, store)
}

// GetMyStores - GET /stores (authenticated)
func GetMyStores(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var stores []models.Store
	// Use subquery to count products
	database.GetDB().
		Model(&models.Store{}).
		Select("stores.*, (SELECT count(*) FROM products WHERE products.store_id = stores.id AND products.deleted_at IS NULL) as product_count").
		Where("user_id = ?", userID).
		Find(&stores)

	c.JSON(http.StatusOK, stores)
}

// UpdateStore - PUT /stores/:store_id
func UpdateStore(c *gin.Context) {
	id := c.Param("store_id")
	userID, _ := c.Get("user_id")

	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var input models.UpdateStoreInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update slug if provided (with uniqueness check)
	if input.Slug != "" && input.Slug != store.Slug {
		var existingStore models.Store
		if database.GetDB().Where("slug = ? AND id != ?", input.Slug, store.ID).First(&existingStore).Error == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Slug sudah digunakan toko lain"})
			return
		}
		store.Slug = input.Slug
	}

	if input.Name != "" {
		store.Name = input.Name
	}
	if input.Whatsapp != "" {
		store.Whatsapp = input.Whatsapp
	}
	store.Description = input.Description
	
	if input.LogoURL != "" {
		fmt.Printf("UPDATING LOGO: %s\n", input.LogoURL)
		store.LogoURL = input.LogoURL
	}
	if input.FaviconURL != "" {
		store.FaviconURL = input.FaviconURL
	}

	// Update HideBranding (Plan check required)
	if input.HideBranding != nil {
		// Fetch user plan
		var user models.User
		if err := database.GetDB().Preload("Plan").First(&user, userID).Error; err == nil && user.Plan != nil {
			if user.Plan.RemoveBranding {
				store.HideBranding = *input.HideBranding
			} else if *input.HideBranding {
				// User trying to enable hide branding without permission
				// We fail silently or just don't set it (UI should prevent this too)
				fmt.Println("Warning: User tried to hide branding without permission")
			} else {
				// User explicitly setting to false (always allowed)
				store.HideBranding = false
			}
		}
	}

	if err := database.GetDB().Save(&store).Error; err != nil {
		fmt.Printf("DB SAVE ERROR: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update store"})
		return
	}
	c.JSON(http.StatusOK, store)
}

// DeleteStore - DELETE /stores/:store_id
func DeleteStore(c *gin.Context) {
	id := c.Param("store_id")
	userID, _ := c.Get("user_id")

	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", id, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	database.GetDB().Delete(&store)
	c.JSON(http.StatusOK, gin.H{"message": "Store deleted"})
}
