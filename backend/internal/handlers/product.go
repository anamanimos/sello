package handlers

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// GetProductsByStore - GET /store/:slug/products (public)
// GetProductsByStore - GET /store/:slug/products (public)
func GetProductsByStore(c *gin.Context) {
	slug := c.Param("slug")
	categoryID := c.Query("category_id")
    search := c.Query("search")

	var store models.Store
	if err := database.GetDB().Where("slug = ?", slug).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	var products []models.Product
    query := database.GetDB().Where("store_id = ?", store.ID)
    
    if search != "" {
		// Case insensitive search using LOWER
		searchLower := strings.ToLower(search)
		searchPattern := "%" + searchLower + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
	}

    if categoryID != "" {
        query = query.Where("category_id = ?", categoryID)
    }

	query.Order("created_at DESC").Find(&products)

	c.JSON(http.StatusOK, gin.H{"data": products})
}

// GetProductsByStoreID - GET /stores/:store_id/products (authenticated)
// GetProductsByStoreID - GET /stores/:store_id/products (authenticated)
// Supports: ?page=1&limit=10&search=keyword&category_id=uuid
func GetProductsByStoreID(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// 1. Verify store belongs to user
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// 2. Parse Query Params
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	search := c.Query("search")
	categoryID := c.Query("category_id")

	// 3. Build Query
	query := database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID)

	if search != "" {
		// Case insensitive search using LOWER and proper grouping
		searchLower := strings.ToLower(search)
		searchPattern := "%" + searchLower + "%"
		query = query.Where("(LOWER(name) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// 4. Count Total
	var total int64
	query.Count(&total)

	// 5. Fetch Data with Pagination
	offset := (page - 1) * limit
	var products []models.Product
	query.Preload("Category").Order("created_at DESC").Limit(limit).Offset(offset).Find(&products)

	// 6. Return Response
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
		},
	})
}

// CreateProduct - POST /stores/:store_id/products
func CreateProduct(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store belongs to user
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Check product limit (Plan Based)
	var user models.User
	if err := database.GetDB().Preload("Plan").First(&user, store.UserID).Error; err == nil {
		maxProducts := 20 // Default
		if user.Plan != nil {
			maxProducts = user.Plan.MaxProductsPerStore
		} else {
			// Fallback system setting
			var sysSetting models.SystemSetting
			if database.GetDB().Where("key = ?", "max_products_per_store").First(&sysSetting).Error == nil {
				fmt.Sscanf(sysSetting.Value, "%d", &maxProducts)
			}
		}

		var productCount int64
		database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)
		if productCount >= int64(maxProducts) {
			c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Batas maksimum produk (%d) tercapai. Upgrade paket Anda.", maxProducts)})
			return
		}
	}

	var input models.CreateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var categoryID *uuid.UUID
	if input.CategoryID != "" {
		if id, err := uuid.Parse(input.CategoryID); err == nil {
			categoryID = &id
		}
	}

	storeUUID, _ := uuid.Parse(storeID)
    
    // Generate Slug
    slug := generateSlug(input.Name)
    // Ensure uniqueness within store
    originalSlug := slug
    counter := 1
    for {
        var count int64
        database.GetDB().Model(&models.Product{}).Where("store_id = ? AND slug = ?", storeUUID, slug).Count(&count)
        if count == 0 {
            break
        }
        slug = fmt.Sprintf("%s-%d", originalSlug, counter)
        counter++
    }

	product := models.Product{
		StoreID:       storeUUID,
		Name:          input.Name,
		Slug:          slug,
		Price:         input.Price,
		DiscountPrice: input.DiscountPrice,
		Description:   input.Description,
		ImageURL:      input.ImageURL,
		ImageSize:     input.ImageSize,
		ExternalLink:  input.ExternalLink,
		CategoryID:    categoryID,
	}

	if err := database.GetDB().Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// GetProduct - GET /products/:id
func GetProduct(c *gin.Context) {
	id := c.Param("id")

	var product models.Product
	if err := database.GetDB().First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct - PUT /products/:id
func UpdateProduct(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Verify ownership
	if product.Store.UserID != uuid.MustParse(userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	var input models.UpdateProductInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Name != "" {
		product.Name = input.Name
        // Regenerate unique slug
        slug := generateSlug(input.Name)
        originalSlug := slug
        counter := 1
        for {
            var count int64
            // Exclude current product
            database.GetDB().Model(&models.Product{}).Where("store_id = ? AND slug = ? AND id != ?", product.StoreID, slug, product.ID).Count(&count)
            if count == 0 {
                break
            }
            slug = fmt.Sprintf("%s-%d", originalSlug, counter)
            counter++
        }
        product.Slug = slug
	}
	if input.Price > 0 {
		product.Price = input.Price
	}
	// DiscountPrice is pointer, so we can check if it's nil (not provided) or has value
	if input.DiscountPrice != nil {
		product.DiscountPrice = input.DiscountPrice
	}
	if input.Description != "" {
		product.Description = input.Description
	}
	if input.ImageURL != nil && *input.ImageURL != "" {
		product.ImageURL = *input.ImageURL
	}
	if input.ExternalLink != nil {
		product.ExternalLink = *input.ExternalLink
	}
	if input.CategoryID != nil {
		if *input.CategoryID == "" {
			product.CategoryID = nil
		} else {
			if id, err := uuid.Parse(*input.CategoryID); err == nil {
				product.CategoryID = &id
			}
		}
	}

	if err := database.GetDB().Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}
	c.JSON(http.StatusOK, product)
}

// DeleteProduct - DELETE /products/:id
func DeleteProduct(c *gin.Context) {
	id := c.Param("id")
	userID, _ := c.Get("user_id")

	var product models.Product
	if err := database.GetDB().Preload("Store").First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Verify ownership
	if product.Store.UserID != uuid.MustParse(userID.(string)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
		return
	}

	// Delete images from Cloudinary
	if cld == nil {
		InitCloudinary()
	}
	
	ctx := context.Background()

	// 1. Delete Primary Image
	if product.ImageURL != "" && strings.Contains(product.ImageURL, "/sello/") {
		// Extract public ID logic (simplified)
		// URL: https://res.cloudinary.com/../sello/slug/id.jpg -> sello/slug/id
		// Ideally we store public_id in DB, but parsing URL works for now
		parts := strings.Split(product.ImageURL, "/sello/")
		if len(parts) > 1 {
			publicID := "sello/" + strings.Split(parts[1], ".")[0]
			cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
		}
	}

	// 2. Delete Gallery Images
	var images []models.ProductImage
	database.GetDB().Where("product_id = ?", product.ID).Find(&images)
	for _, img := range images {
		if img.ImageURL != "" && strings.Contains(img.ImageURL, "/sello/") {
			parts := strings.Split(img.ImageURL, "/sello/")
			if len(parts) > 1 {
				publicID := "sello/" + strings.Split(parts[1], ".")[0]
				cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
			}
		}
	}

	database.GetDB().Delete(&product)
	c.JSON(http.StatusOK, gin.H{"message": "Product and images deleted"})
}
