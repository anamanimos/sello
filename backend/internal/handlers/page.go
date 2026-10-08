package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"fmt"
	"sello/internal/database"
	"sello/internal/models"
)

// RenderStorePage - GET /s/:slug (HTML page)
// RenderStorePage - GET /s/:slug (HTML page)
func RenderStorePage(c *gin.Context) {
	slug := c.Param("slug")
	search := c.Query("q")
	categoryID := c.Query("cat")

	var store models.Store
	result := database.GetDB().Where("slug = ?", slug).First(&store)

	if result.Error != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{
			"message": "Toko tidak ditemukan",
		})
		return
	}

	// Check if branding should be shown
	showBranding := true
	var owner models.User
	if err := database.GetDB().Where("id = ?", store.UserID).Preload("Plan").First(&owner).Error; err == nil {
		// Branding hidden only if Store setting is ON AND Plan allows it
		if store.HideBranding && owner.Plan != nil && owner.Plan.RemoveBranding {
			showBranding = false
		}
		// DEBUG
		fmt.Printf("DEBUG_BRANDING: Slug=%s Hide=%t PlanRemove=%t -> SHOW=%t\n", slug, store.HideBranding, (owner.Plan != nil && owner.Plan.RemoveBranding), showBranding)
	}

	// Fetch Categories
	var categories []models.Category
	database.GetDB().Where("store_id = ?", store.ID).Order("name ASC").Find(&categories)

	// Fetch Products with Filter
	// Fetch Products with Filter
	var products []models.Product
	query := database.GetDB().Where("store_id = ?", store.ID)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	query.Order("created_at DESC").Find(&products)

	// TRACK VISIT
	TrackEvent(c, store.ID, "view_store", nil)

	c.HTML(http.StatusOK, "store.html", gin.H{
		"Store":        store,
		"Products":     products,
		"Categories":   categories,
		"SearchQuery":  search,
		"SelectedCat":  categoryID,
		"ShowBranding": showBranding,
	})
}

// RenderProductDetail - GET /s/:slug/p/:product_slug
func RenderProductDetail(c *gin.Context) {
	slug := c.Param("slug")
	productParam := c.Param("id") // Changed routing to :id -> :product_slug in main.go, but let's keep it generic or rename param

	var store models.Store
	if err := database.GetDB().Where("slug = ?", slug).First(&store).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Toko tidak ditemukan"})
		return
	}

	var product models.Product
    // Try to parse as UUID
    if id, err := uuid.Parse(productParam); err == nil {
         // It is UUID
         if err := database.GetDB().Where("id = ? AND store_id = ?", id, store.ID).Preload("Images").First(&product).Error; err != nil {
             c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Produk tidak ditemukan"})
             return
         }
    } else {
        // It is Slug
         if err := database.GetDB().Where("slug = ? AND store_id = ?", productParam, store.ID).Preload("Images").First(&product).Error; err != nil {
             c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Produk tidak ditemukan"})
             return
         }
    }

	// TRACK PRODUCT VIEW
	TrackEvent(c, store.ID, "view_product", &product.ID)

	c.HTML(http.StatusOK, "product_detail.html", gin.H{
		"Store":   store,
		"Product": product,
	})
}

// RenderStoreFromDomain - handles store access via subdomain or custom domain
// Used when middleware sets store_slug in context
func RenderStoreFromDomain(c *gin.Context) {
	slugVal, exists := c.Get("store_slug")
	if !exists {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Toko tidak ditemukan"})
		return
	}
	slug := slugVal.(string)
	
	search := c.Query("q")
	categoryID := c.Query("cat")

	var store models.Store
	if err := database.GetDB().Where("slug = ?", slug).First(&store).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Toko tidak ditemukan"})
		return
	}

	if store.IsSuspended {
		c.HTML(http.StatusForbidden, "store_suspended.html", gin.H{
			"Reason": store.SuspensionReason,
		})
		return
	}

	// Fetch Categories
	var categories []models.Category
	database.GetDB().Where("store_id = ?", store.ID).Order("name ASC").Find(&categories)

	// Fetch Products with Filter
	var products []models.Product
	query := database.GetDB().Where("store_id = ?", store.ID)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	// Check if branding should be shown
	showBranding := true
	var owner models.User
	if err := database.GetDB().Where("id = ?", store.UserID).Preload("Plan").First(&owner).Error; err == nil {
		// Branding hidden only if Store setting is ON AND Plan allows it
		if store.HideBranding && owner.Plan != nil && owner.Plan.RemoveBranding {
			showBranding = false
		}
	}

	query.Order("created_at DESC").Find(&products)

	// Track Visit
	TrackEvent(c, store.ID, "view_store", nil)
    
    // Determine isSubdomain context for link generation
    // Better check: AccessType
    accessType, _ := c.Get("access_type")
    isSubdomainCtx := accessType == "subdomain" || accessType == "custom_domain"

	// TRY THEME RENDERER
	html := renderStoreWithTheme(store, products, categories, "homepage", isSubdomainCtx, owner)
	if html != "" {
		html = injectBranding(html, store)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	c.HTML(http.StatusOK, "store.html", gin.H{
		"Store":        store,
		"Products":     products,
		"Categories":   categories,
		"SearchQuery":  search,
		"SelectedCat":  categoryID,
		"IsSubdomain":  isSubdomainCtx, // Use calculated flag
		"ShowBranding": showBranding,
	})
}

// RenderProductFromDomain - handles product detail access via subdomain or custom domain
func RenderProductFromDomain(c *gin.Context) {
	slugVal, exists := c.Get("store_slug")
	if !exists {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Toko tidak ditemukan"})
		return
	}
	slug := slugVal.(string)
	productParam := c.Param("id")

	var store models.Store
	if err := database.GetDB().Where("slug = ?", slug).First(&store).Error; err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Toko tidak ditemukan"})
		return
	}

	if store.IsSuspended {
		c.HTML(http.StatusForbidden, "store_suspended.html", gin.H{
			"Reason": store.SuspensionReason,
		})
		return
	}

	var product models.Product
    if id, err := uuid.Parse(productParam); err == nil {
         if err := database.GetDB().Where("id = ? AND store_id = ?", id, store.ID).Preload("Images").First(&product).Error; err != nil {
            c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Produk tidak ditemukan"})
            return
         }
    } else {
         if err := database.GetDB().Where("slug = ? AND store_id = ?", productParam, store.ID).Preload("Images").First(&product).Error; err != nil {
            c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Produk tidak ditemukan"})
            return
         }
    }

	// Track Product View
	TrackEvent(c, store.ID, "view_product", &product.ID)

	// Logic Buy URL
	buyURL := ""
	if product.ExternalLink != "" {
		buyURL = product.ExternalLink
	} else {
		buyURL = fmt.Sprintf("https://wa.me/%s?text=Halo, saya tertarik dengan %s (Rp %d)",
			store.Whatsapp,
			product.Name,
			product.Price,
		)
	}

	// Fetch Store Owner for Fallback
	var owner models.User
	database.GetDB().Where("id = ?", store.UserID).First(&owner)

	// TRY THEME RENDERER
	html := renderProductWithTheme(store, product, buyURL, owner, true)
	if html != "" {
		html = injectBranding(html, store)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	c.HTML(http.StatusOK, "product_detail.html", gin.H{
		"Store":       store,
		"Product":     product,
		"IsSubdomain": true,
	})
}
