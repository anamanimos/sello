package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
)

var cld *cloudinary.Cloudinary

// InitCloudinary initializes the Cloudinary client
func InitCloudinary() error {
	var err error
	cld, err = cloudinary.NewFromParams(
		config.AppConfig.CloudinaryCloudName,
		config.AppConfig.CloudinaryAPIKey,
		config.AppConfig.CloudinaryAPISecret,
	)
	return err
}

// UploadImageFromURL uploads an image from a URL to Cloudinary
// Used by seeder to upload Unsplash images
func UploadImageFromURL(imageURL, storeSlug, productSlug string) (secureURL string, publicID string, err error) {
	// Initialize Cloudinary if not already
	if cld == nil {
		if err := InitCloudinary(); err != nil {
			return "", "", fmt.Errorf("failed to init cloudinary: %v", err)
		}
	}

	if cld == nil {
		return "", "", fmt.Errorf("cloudinary client is nil")
	}

	// Generate unique public ID
	pubID := fmt.Sprintf("sello/%s/%s_%d", storeSlug, productSlug, time.Now().UnixNano())

	ctx := context.Background()
	uploadResult, err := cld.Upload.Upload(ctx, imageURL, uploader.UploadParams{
		PublicID:       pubID,
		Folder:         "sello/" + storeSlug,
		ResourceType:   "image",
		Transformation: "c_limit,w_1200,h_1200,q_auto:good",
	})

	if err != nil {
		return "", "", fmt.Errorf("upload failed: %v", err)
	}

	return uploadResult.SecureURL, uploadResult.PublicID, nil
}

// UploadImage - POST /api/stores/:store_id/upload
// Uploads an image to Cloudinary under the store's folder
func UploadImage(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Check storage limit (Plan Based)
	var user models.User
	var maxStorageMB int = 50 // Default 50MB
	if err := database.GetDB().Preload("Plan").First(&user, userID).Error; err == nil {
		if user.Plan != nil {
			maxStorageMB = user.Plan.MaxStorageMB
		}
	}
	limitBytes := int64(maxStorageMB) * 1024 * 1024

	var totalSize int64
	database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID).Select("COALESCE(SUM(image_size), 0)").Scan(&totalSize)
	if totalSize >= limitBytes {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Storage limit exceeded (%dMB)", maxStorageMB)})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No image file provided"})
		return
	}
	defer file.Close()

	// Validate file type
	ext := strings.ToLower(filepath.Ext(header.Filename))
	validExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !validExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type. Allowed: jpg, jpeg, png, gif, webp"})
		return
	}

	// Validate file size (max 5MB per image)
	if header.Size > 5*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Max 5MB per image"})
		return
	}

	// Check total storage won't exceed limit
	if totalSize+header.Size > limitBytes {
		c.JSON(http.StatusForbidden, gin.H{
			"error":        "Upload would exceed storage limit",
			"storage_used": totalSize,
			"storage_max":  limitBytes,
		})
		return
	}

	// CLOUDINARY UPLOAD (Restored as per user request)
	
	// Initialize Cloudinary if not already (safeguard)
	if cld == nil {
		if err := InitCloudinary(); err != nil {
			log.Printf("Error initializing Cloudinary: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize image service: " + err.Error()})
			return
		}
	}

	// Double check cld is not nil
	if cld == nil {
		log.Println("Error: Cloudinary client is nil after init")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary client initialization failed"})
		return
	}

	// Generate unique public ID
	publicID := fmt.Sprintf("sello/%s/%s_%d", store.Slug, uuid.New().String()[:8], time.Now().Unix())
	
	// Log upload attempt
	log.Printf("Attempting upload to Cloudinary. PublicID: %s", publicID)

	ctx := context.Background()
	// Use explicit deferred recovery for this handler logic just in case SDK panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic in Cloudinary Upload: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error during upload"})
		}
	}()

	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		Folder:         "sello/" + store.Slug,
		ResourceType:   "image",
		Transformation: "c_limit,w_1200,h_1200,q_auto:good",
	})

	if err != nil {
		log.Printf("Cloudinary Upload Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image: " + err.Error()})
		return
	}

	log.Printf("Upload Success: %s", uploadResult.SecureURL)

	c.JSON(http.StatusOK, gin.H{
		"url":        uploadResult.SecureURL,
		"public_id":  uploadResult.PublicID,
		"width":      uploadResult.Width,
		"height":     uploadResult.Height,
		"size":       header.Size,
		"format":     uploadResult.Format,
		"store_slug": store.Slug,
	})

	/*
	// Initialize Cloudinary if not already
	if cld == nil {
		if err := InitCloudinary(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize image service"})
			return
		}
	}
	// ... (Cloudinary code commented out or removed)
	*/

}

// DeleteImage - DELETE /api/images/:public_id
// Deletes an image from Cloudinary
func DeleteImage(c *gin.Context) {
	publicID := c.Query("public_id")
	if publicID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "public_id is required"})
		return
	}

	userID, _ := c.Get("user_id")

	// Verify the image belongs to user's store (check public_id contains their store slug)
	var stores []models.Store
	database.GetDB().Where("user_id = ?", userID).Find(&stores)

	authorized := false
	for _, store := range stores {
		if strings.Contains(publicID, store.Slug) {
			authorized = true
			break
		}
	}

	if !authorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this image"})
		return
	}

	// Initialize Cloudinary if not already
	if cld == nil {
		if err := InitCloudinary(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initialize image service"})
			return
		}
	}

	// Delete from Cloudinary
	ctx := context.Background()
	_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}

// GetStorageUsage - GET /api/stores/:store_id/storage
func GetStorageUsage(c *gin.Context) {
	storeID := c.Param("store_id")
	userID, _ := c.Get("user_id")

	// Verify store ownership
	var store models.Store
	if err := database.GetDB().Where("id = ? AND user_id = ?", storeID, userID).First(&store).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Calculate total storage used
	var totalSize int64
	database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID).Select("COALESCE(SUM(image_size), 0)").Scan(&totalSize)

	// Count products
	var productCount int64
	database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)

	// Get User Plan Limits
	var user models.User
	maxStorageMB := 50
	maxProducts := 20
	if err := database.GetDB().Preload("Plan").First(&user, userID).Error; err == nil && user.Plan != nil {
		maxStorageMB = user.Plan.MaxStorageMB
		maxProducts = user.Plan.MaxProductsPerStore
	}
	maxStorageBytes := int64(maxStorageMB) * 1024 * 1024

	c.JSON(http.StatusOK, gin.H{
		"storage_used":    totalSize,
		"storage_max":     maxStorageBytes,
		"storage_percent": float64(totalSize) / float64(maxStorageBytes) * 100,
		"product_count":   productCount,
		"product_max":     maxProducts,
	})
}
