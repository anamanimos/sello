package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// GetProductImages - GET /products/:id/images
func GetProductImages(c *gin.Context) {
	productID := c.Param("id")

	var images []models.ProductImage
	database.GetDB().Where("product_id = ?", productID).Order("sort_order ASC").Find(&images)

	c.JSON(http.StatusOK, images)
}

// AddProductImage - POST /products/:id/images
func AddProductImage(c *gin.Context) {
	productID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify product ownership
	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	userIDStr, _ := userID.(string)
	if product.Store.UserID.String() != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Recovery for this specific handler
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic in AddProductImage:", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Internal Panic: %v", r)})
		}
	}()

	var input struct {
		ImageURL  string `json:"image_url" binding:"required"`
		ImageSize int64  `json:"image_size"`
		IsPrimary bool   `json:"is_primary"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Calculate sort order based on existing images count
	var count int64
	database.GetDB().Model(&models.ProductImage{}).Where("product_id = ?", productID).Count(&count)
	maxOrder := int(count)

	productUUID, _ := uuid.Parse(productID)
	image := models.ProductImage{
		ID:        uuid.New(),
		ProductID: productUUID,
		ImageURL:  input.ImageURL,
		ImageSize: input.ImageSize,
		IsPrimary: input.IsPrimary,
		SortOrder: maxOrder + 1,
	}

	// If this is primary, unset other primaries
	if input.IsPrimary {
		database.GetDB().Model(&models.ProductImage{}).
			Where("product_id = ?", productID).
			Update("is_primary", false)

		// Also update product's main image_url
		product.ImageURL = input.ImageURL
		database.GetDB().Save(&product)
	}

	if err := database.GetDB().Create(&image).Error; err != nil {
		fmt.Println("Error creating product image:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add image: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, image)
}

// SetPrimaryImage - PUT /products/:id/images/:image_id/primary
func SetPrimaryImage(c *gin.Context) {
	productID := c.Param("id")
	imageID := c.Param("image_id")
	userID, _ := c.Get("user_id")

	// Verify ownership
	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	userIDStr, _ := userID.(string)
	if product.Store.UserID.String() != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Get the image
	var image models.ProductImage
	if err := database.GetDB().First(&image, "id = ? AND product_id = ?", imageID, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Unset all primaries
	database.GetDB().Model(&models.ProductImage{}).
		Where("product_id = ?", productID).
		Update("is_primary", false)

	// Set this as primary
	image.IsPrimary = true
	database.GetDB().Save(&image)

	// Update product's main image
	product.ImageURL = image.ImageURL
	database.GetDB().Save(&product)

	c.JSON(http.StatusOK, gin.H{"message": "Primary image updated"})
}

// ReorderImages - PUT /products/:id/images/reorder
func ReorderImages(c *gin.Context) {
	productID := c.Param("id")
	userID, _ := c.Get("user_id")

	// Verify ownership
	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	userIDStr, _ := userID.(string)
	if product.Store.UserID.String() != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var input struct {
		ImageIDs []string `json:"image_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update sort orders
	for i, id := range input.ImageIDs {
		database.GetDB().Model(&models.ProductImage{}).
			Where("id = ? AND product_id = ?", id, productID).
			Update("sort_order", i)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Images reordered"})
}

// DeleteProductImage - DELETE /products/:id/images/:image_id
func DeleteProductImage(c *gin.Context) {
	productID := c.Param("id")
	imageID := c.Param("image_id")
	userID, _ := c.Get("user_id")

	// Verify ownership
	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	userIDStr, _ := userID.(string)
	if product.Store.UserID.String() != userIDStr {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Delete the image
	var image models.ProductImage
	if err := database.GetDB().First(&image, "id = ? AND product_id = ?", imageID, productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	wasPrimary := image.IsPrimary
	database.GetDB().Delete(&image)

	// If was primary, set first remaining as primary
	if wasPrimary {
		var firstImage models.ProductImage
		if database.GetDB().Where("product_id = ?", productID).Order("sort_order ASC").First(&firstImage).Error == nil {
			firstImage.IsPrimary = true
			database.GetDB().Save(&firstImage)
			product.ImageURL = firstImage.ImageURL
			database.GetDB().Save(&product)
		} else {
			// No images left
			product.ImageURL = ""
			database.GetDB().Save(&product)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted"})
}
