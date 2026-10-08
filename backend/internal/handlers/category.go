package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// GetStoreCategories - GET /api/stores/:store_id/categories
func GetStoreCategories(c *gin.Context) {
	storeID := c.Param("store_id")

	var categories []models.Category
	if err := database.GetDB().Where("store_id = ?", storeID).Order("name asc").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCategory - POST /api/stores/:store_id/categories
func CreateCategory(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var input models.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate Slug
	slug := models.GenerateSlug(input.Name)

	// Check duplicates in store
	var count int64
	database.GetDB().Model(&models.Category{}).Where("store_id = ? AND slug = ?", storeID, slug).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Category already exists"})
		return
	}

	storeUUID, _ := uuid.Parse(storeID)
	category := models.Category{
		StoreID: storeUUID,
		Name:    input.Name,
		Slug:    slug,
	}

	if err := database.GetDB().Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// UpdateCategory - PUT /api/categories/:id
func UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var category models.Category
	if err := database.GetDB().Preload("Store").First(&category, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Verify ownership via Store
	if category.Store.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var input models.UpdateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.Name = input.Name
	category.Slug = models.GenerateSlug(input.Name) // Update slug too? usually yes

	database.GetDB().Save(&category)
	c.JSON(http.StatusOK, category)
}

// DeleteCategory - DELETE /api/categories/:id
func DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var category models.Category
	if err := database.GetDB().Preload("Store").First(&category, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if category.Store.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Set products category_id to null
	database.GetDB().Model(&models.Product{}).Where("category_id = ?", category.ID).Update("category_id", nil)

	database.GetDB().Delete(&category)
	c.JSON(http.StatusOK, gin.H{"message": "Category deleted"})
}
