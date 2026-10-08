package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	// Added time
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// ThemeRenderer - Helper untuk render template tema
type ThemeRenderer struct {
	Store      models.Store
	Products   []models.Product
	Product    *models.Product
	Categories []models.Category
	Config     models.ThemeConfig
	Owner      models.User
}

// RenderThemedStorePage - GET /s/:slug (dengan tema)
func RenderThemedStorePage(c *gin.Context) {
	slug := c.Param("slug")
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

	query.Order("created_at DESC").Find(&products)

	// TRACK VISIT
	TrackEvent(c, store.ID, "view_store", nil)

	// Fetch Store Owner for Fallback
	var owner models.User
	database.GetDB().Where("id = ?", store.UserID).First(&owner)

	// Get store's active theme
	html := renderStoreWithTheme(store, products, categories, "homepage", false, owner)

	if html != "" {
		html = injectBranding(html, store)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	// Fallback to default template
	c.HTML(http.StatusOK, "store.html", gin.H{
		"Store":       store,
		"Products":    products,
		"Categories":  categories,
		"SearchQuery": search,
		"SelectedCat": categoryID,
	})
}

// RenderThemedProductDetail - GET /s/:slug/p/:id (dengan tema)
func RenderThemedProductDetail(c *gin.Context) {
	slug := c.Param("slug")
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

	// TRACK PRODUCT VIEW
	TrackEvent(c, store.ID, "view_product", &product.ID)

	// LOGIC: Determine Buy URL (Standardized Backend Logic)
	buyURL := ""
	if product.ExternalLink != "" {
		buyURL = product.ExternalLink
	} else {
		// Default to WhatsApp
		buyURL = fmt.Sprintf("https://wa.me/%s?text=Halo, saya tertarik dengan %s (Rp %d)", 
			store.Whatsapp, 
			product.Name, 
			product.Price,
		)
	}

	// Fetch Store Owner for Fallback
	var owner models.User
	database.GetDB().Where("id = ?", store.UserID).First(&owner)

	// Get store's active theme
	html := renderProductWithTheme(store, product, buyURL, owner, false)

	if html != "" {
		html = injectBranding(html, store)
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}

	// Fallback to default template
	c.HTML(http.StatusOK, "product_detail.html", gin.H{
		"Store":   store,
		"Product": product,
	})
}

// Helper function (Local to package or use shared from theme.go if exported, but let's redefine simplified version to avoid cyclic dep if any)
// Actually, handlers package is same.
// We need to update getStoreThemeAndConfig in theme_renderer.go to use the new Logic.

// getStoreThemeAndConfig (Updated for Disk Schema)
func getStoreThemeAndConfig(storeID string) (*models.Theme, *models.ThemeMeta, map[string]interface{}) {
	var storeTheme models.StoreTheme
	err := database.GetDB().Where("store_id = ? AND is_active = ?", storeID, true).First(&storeTheme).Error

	var theme *models.Theme
    var meta models.ThemeMeta
    config := make(map[string]interface{})

	if err != nil {
		// Try default theme "starter"
		t, err := loadThemeFromDisk("starter")
		if err == nil {
			theme = t
		} else {
			return nil, nil, nil
		}
	} else {
		// Load from disk using ThemeName
		t, err := loadThemeFromDisk(storeTheme.ThemeName)
		if err != nil {
			// Fallback to starter if missing
			t, _ = loadThemeFromDisk("starter")
		}
		theme = t
	}
	
	if theme == nil {
		return nil, nil, nil
	}

	// 1. Read ThemeMeta
    // theme.DirectoryPath is already set by loadThemeFromDisk
	meta = readDiskThemeMeta(theme.DirectoryPath)

    // 2. Populate defaults
    for _, setting := range meta.Settings {
        if setting.ID != "" {
            config[setting.ID] = setting.Default
        }
		// Handle sections defaults too if needed, but for now simple
    }

    // 3. Merge User Custom Config
    if storeTheme.ID != uuid.Nil {
        var userConfig map[string]interface{}
        if len(storeTheme.CustomConfig) > 0 {
             json.Unmarshal(storeTheme.CustomConfig, &userConfig)
        }
        for k, v := range userConfig {
            config[k] = v
        }
    }

	return theme, &meta, config
}

// RenderThemePreviewPage - GET /dashboard/theme/preview/:slug
func RenderThemePreviewPage(c *gin.Context) {
	themeSlug := c.Param("slug")

	// Dummy Store Data
	dummyStore := models.Store{
		ID:          uuid.New(),
		Name:        "Toko Demo Sello",
		Slug:        themeSlug, // Using theme slug mostly for context, but store slug is usually different
		Description: "Ini adalah contoh toko untuk preview tema.",
		LogoURL:     "", // Empty to trigger default logo in theme
		FaviconURL:  "/static/images/sello-icon.ico",
	}

	// Dummy Categories
	dummyCategories := []models.Category{
		{ID: uuid.New(), Name: "Pakaian", Slug: "pakaian"},
		{ID: uuid.New(), Name: "Elektronik", Slug: "elektronik"},
		{ID: uuid.New(), Name: "Aksesoris", Slug: "aksesoris"},
	}

	// Dummy Products
	dummyProducts := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Kamera Mirrorless Pro",
			Description: "Kamera mirrorless dengan sensor full-frame 24MP, video 4K, dan autofocus cepat. Cocok untuk profesional.",
			Price:       15000000,
			ImageURL:    "https://images.unsplash.com/photo-1516035069371-29a1b244cc32?auto=format&fit=crop&w=400&q=80",
		},
		{
			ID:            uuid.New(),
			Name:          "Headphone Noise Cancelling",
			Description:   "Headphone nirkabel dengan fitur active noise cancelling terbaik di kelasnya.",
			Price:         3500000,
			DiscountPrice: func() *int64 { i := int64(2999000); return &i }(),
			ImageURL:      "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=400&q=80",
		},
		{
			ID:          uuid.New(),
			Name:        "Smartwatch Series 5",
			Description: "Jam tangan pintar dengan monitor detak jantung, GPS, dan tahan air hingga 50 meter.",
			Price:       4200000,
			ImageURL:    "https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=400&q=80",
		},
		{
			ID:          uuid.New(),
			Name:        "Sneakers Urban Run",
			Description: "Sepatu lari yang nyaman dengan desain modern untuk gaya hidup perkotaan.",
			Price:       1200000,
			ImageURL:    "https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=400&q=80",
		},
	}

	// We need to fetch the theme explicitly since it's not linked to a store via DB
	theme, err := loadThemeFromDisk(themeSlug)
	if err != nil {
		c.HTML(http.StatusNotFound, "404.html", gin.H{"message": "Tema tidak ditemukan"})
		return
	}

	// Manually construct HTML using the render helper but providing the specific theme
	// Note: renderStoreWithTheme normally fetches theme from DB via store ID.
	// We need a variant that accepts the theme object directly.
	
	// Dummy Owner for Preview
	dummyOwner := models.User{
		AvatarURL: "",
	}

	html := renderPreviewWithTheme(dummyStore, dummyProducts, dummyCategories, theme, dummyOwner)
	
	if html != "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
		return
	}
	
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render preview"})
}

// renderPreviewWithTheme is like renderStoreWithTheme but takes an explicit theme object
func renderPreviewWithTheme(store models.Store, products []models.Product, categories []models.Category, theme *models.Theme, owner models.User) string {
    // 1. Get Config (Default only)
    meta := readDiskThemeMeta(theme.DirectoryPath)
    config := make(models.ThemeConfig)
    for _, setting := range meta.Settings {
        if setting.ID != "" {
            config[setting.ID] = setting.Default
        }
    }

	templateFile := "home.html"
    if meta.Templates != nil {
        if val, ok := meta.Templates["index"]; ok {
            templateFile = val
        }
    }

	path := filepath.Join("web", "themes", theme.DirectoryPath, templateFile)
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	html := string(content)

	html = replaceStorePlaceholders(html, store, config, owner)
	html = replaceConfigPlaceholders(html, config)
	html = replaceProductsLoop(html, products, store, false)
	html = replaceCategoriesLoop(html, categories, store)
	html = strings.ReplaceAll(html, "{{products_count}}", strconv.Itoa(len(products)))

	return html
}

func renderStoreWithTheme(store models.Store, products []models.Product, categories []models.Category, pageType string, isSubdomain bool, owner models.User) string {
	theme, meta, config := getStoreThemeAndConfig(store.ID.String())
	if theme == nil {
		return ""
	}

	templateFile := "home.html"
    if meta != nil && meta.Templates != nil {
        if val, ok := meta.Templates["index"]; ok && pageType == "homepage" {
            templateFile = val
        }
    }

	path := filepath.Join("web", "themes", theme.DirectoryPath, templateFile)
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	html := string(content)

	html = replaceStorePlaceholders(html, store, config, owner)
	html = replaceConfigPlaceholders(html, config)
	html = replaceProductsLoop(html, products, store, isSubdomain) // Updated
	html = replaceCategoriesLoop(html, categories, store)
	html = strings.ReplaceAll(html, "{{products_count}}", strconv.Itoa(len(products)))

	return html
}

func renderProductWithTheme(store models.Store, product models.Product, buyURL string, owner models.User, isSubdomain bool) string {
	theme, _, config := getStoreThemeAndConfig(store.ID.String())
	if theme == nil {
		return ""
	}

	templateName := "product.html"
	path := filepath.Join("web", "themes", theme.DirectoryPath, templateName)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ERROR: Failed to read theme file %s: %v", path, err)
		return ""
	}
	htmlContent := string(content)

	htmlContent = replaceConfigPlaceholders(htmlContent, config)
	htmlContent = replaceProductPlaceholders(htmlContent, product)
	htmlContent = replaceStorePlaceholders(htmlContent, store, config, owner)
	htmlContent = replaceProductImagesLoop(htmlContent, product)
	htmlContent = strings.ReplaceAll(htmlContent, "{{product.buy_url}}", buyURL)

	// Inject Home URL logic
	homeURL := "/s/" + store.Slug
	if isSubdomain {
		homeURL = "/"
	}
	htmlContent = strings.ReplaceAll(htmlContent, "{{home_url}}", homeURL)

	return htmlContent
}

// Helper functions for placeholder replacement
func replaceStorePlaceholders(html string, store models.Store, config map[string]interface{}, owner models.User) string {
	// Determine Logo URL with Priority:
	// 1. Theme Config Logo
	// 2. Store Logo
	// 3. User Avatar
	// 4. Default Fallback
	logoURL := "/static/images/sello-pp.webp" // Default

	if owner.AvatarURL != "" {
		logoURL = owner.AvatarURL
	}
	if store.LogoURL != "" {
		logoURL = store.LogoURL
	}
	// Check theme config
	if val, ok := config["logo"]; ok && val != nil {
		if strVal, ok := val.(string); ok && strVal != "" {
			logoURL = strVal
		}
	}

	html = strings.ReplaceAll(html, "{{store.name}}", store.Name)
	html = strings.ReplaceAll(html, "{{store.description}}", store.Description)
	html = strings.ReplaceAll(html, "{{store.whatsapp}}", store.Whatsapp)
	html = strings.ReplaceAll(html, "{{store.slug}}", store.Slug)
	html = strings.ReplaceAll(html, "{{store.logo_url}}", logoURL)
	html = strings.ReplaceAll(html, "{{store.favicon_url}}", store.FaviconURL)

	// Handle conditionals
	// For logo, we always have a value (fallback), so we set it to true to render the img tag
	// The template usually is: {{#if store.logo_url}}{{store.logo_url}}{{else}}fallback{{/if}}
	// Since we injected the fallback into the variable itself, we can just say True.
	html = processConditional(html, "store.logo_url", true) 
	html = processConditional(html, "store.favicon_url", store.FaviconURL != "")
	html = processConditional(html, "store.description", store.Description != "")

	return html
}

func replaceConfigPlaceholders(html string, config map[string]interface{}) string {
    // Iterate over config map
	for key, value := range config {
        // Handle Arrays (Repeater)
        if items, ok := value.([]interface{}); ok {
             html = processRepeaterLoop(html, key, items)
             continue
        }

        // Regular value
        strVal := fmt.Sprintf("%v", value)
        
        // Prioritize Triple Braces {{{key}}} to avoid leaving outer braces
		html = strings.ReplaceAll(html, fmt.Sprintf("{{{%s}}}", key), strVal)
		html = strings.ReplaceAll(html, fmt.Sprintf("{{%s}}", key), strVal)
        
        // Handle conditional {{#if key}}...{{/if}}
        isTruthy := strVal != "" && strVal != "0" && strVal != "false" && value != nil
        html = processConditional(html, key, isTruthy)
	}

	return html
}

func processRepeaterLoop(html string, key string, items []interface{}) string {
    // Find {{#each key}} ... {{/each}}
    pattern := fmt.Sprintf(`(?s)\{\{#each %s\}\}([\s\S]*?)\{\{/each\}\}`, regexp.QuoteMeta(key))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return html
	}

	itemTemplate := matches[1]
	var resultHTML strings.Builder

	for _, item := range items {
        // item should be map[string]interface{}
        itemMap, ok := item.(map[string]interface{})
        if !ok { continue }
        
        // Render item
        render := itemTemplate
        
        // First, process conditionals WITH else for this item ({{#if this.field}}...{{else}}...{{/if}})
        for k, v := range itemMap {
            // Pattern with else clause
            condWithElsePattern := fmt.Sprintf(`(?s)\{\{#if this\.%s\}\}([\s\S]*?)\{\{else\}\}([\s\S]*?)\{\{/if\}\}`, regexp.QuoteMeta(k))
            condWithElseRe := regexp.MustCompile(condWithElsePattern)
            
            // Check if value is truthy
            valStr := fmt.Sprintf("%v", v)
            isTruthy := valStr != "" && valStr != "0" && valStr != "false" && valStr != "<nil>" && v != nil
            
            // Handle if-else pattern
            if condWithElseRe.MatchString(render) {
                if isTruthy {
                    // Keep the true block (group 1)
                    render = condWithElseRe.ReplaceAllString(render, "$1")
                } else {
                    // Keep the else block (group 2)
                    render = condWithElseRe.ReplaceAllString(render, "$2")
                }
            } else {
                // Handle simple if pattern (no else)
                condPattern := fmt.Sprintf(`(?s)\{\{#if this\.%s\}\}([\s\S]*?)\{\{/if\}\}`, regexp.QuoteMeta(k))
                condRe := regexp.MustCompile(condPattern)
                
                if isTruthy {
                    // Keep the content inside the conditional
                    render = condRe.ReplaceAllString(render, "$1")
                } else {
                    // Remove the entire conditional block
                    render = condRe.ReplaceAllString(render, "")
                }
            }
        }
        
        // Clean up any remaining unprocessed conditionals for this.* fields (with else)
        remainingCondElseRe := regexp.MustCompile(`(?s)\{\{#if this\.[^}]+\}\}[\s\S]*?\{\{else\}\}[\s\S]*?\{\{/if\}\}`)
        render = remainingCondElseRe.ReplaceAllString(render, "")
        
        // Clean up any remaining unprocessed conditionals for this.* fields (without else)
        remainingCondRe := regexp.MustCompile(`(?s)\{\{#if this\.[^}]+\}\}[\s\S]*?\{\{/if\}\}`)
        render = remainingCondRe.ReplaceAllString(render, "")
        
        // Then replace values
        for k, v := range itemMap {
            valStr := fmt.Sprintf("%v", v)
            render = strings.ReplaceAll(render, fmt.Sprintf("{{this.%s}}", k), valStr)
        }
        resultHTML.WriteString(render)
	}

	return re.ReplaceAllString(html, resultHTML.String())
}



func replaceProductPlaceholders(html string, product models.Product) string {
	html = strings.ReplaceAll(html, "{{product.name}}", product.Name)
	html = strings.ReplaceAll(html, "{{product.name}}", product.Name)
	
	// Handle description: if plain text (no tags), convert newlines to <br>
	description := product.Description
	if !strings.Contains(description, "<") && strings.Contains(description, "\n") {
		description = strings.ReplaceAll(description, "\n", "<br>")
	}
	html = strings.ReplaceAll(html, "{{product.description}}", description)

	html = strings.ReplaceAll(html, "{{product.image_url}}", product.ImageURL)
	html = strings.ReplaceAll(html, "{{product.id}}", product.ID.String())
    html = strings.ReplaceAll(html, "{{product.slug}}", product.Slug) // Expose Slug
	html = strings.ReplaceAll(html, "{{product.external_link}}", product.ExternalLink)

	// Handle discount (DiscountPrice is the sale price, show original Price as strikethrough)
	hasDiscount := product.DiscountPrice != nil && *product.DiscountPrice > 0 && *product.DiscountPrice < product.Price
	if hasDiscount {
		html = strings.ReplaceAll(html, "{{product.original_price}}", formatPrice(product.Price))
		html = strings.ReplaceAll(html, "{{product.price}}", formatPrice(*product.DiscountPrice))
		discount := int((float64(product.Price-*product.DiscountPrice) / float64(product.Price)) * 100)
		html = strings.ReplaceAll(html, "{{product.discount_percent}}", strconv.Itoa(discount))
	} else {
        // No discount: price is just price, original_price is empty
        html = strings.ReplaceAll(html, "{{product.price}}", formatPrice(product.Price))
    }
	html = processConditional(html, "product.original_price", hasDiscount)

	// Handle description
	html = processConditional(html, "product.description", product.Description != "")

	// Handle External Link Conditional
	html = processConditional(html, "product.external_link", product.ExternalLink != "")
	
	// Handle Image Conditional
	html = processConditional(html, "product.image_url", product.ImageURL != "")

	return html
}

func replaceProductsLoop(html string, products []models.Product, store models.Store, isSubdomain bool) string {
	log.Println("DEBUG: replaceProductsLoop V2 START")
	// Find the {{#each products}} ... {{/each}} block
	// Use greedy matching (.*) for the body to ensure we don't stop at an inner {{else}}
	re := regexp.MustCompile(`(?s)\{\{#each products\}\}(.*)\{\{else\}\}(.*)\{\{/each\}\}`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 3 {
		log.Println("DEBUG: Falling back to non-else regex")
		// Try without else
		re = regexp.MustCompile(`\{\{#each products\}\}([\s\S]*?)\{\{/each\}\}`)
		matches = re.FindStringSubmatch(html)
		if len(matches) < 2 {
			log.Println("DEBUG: No products loop found")
			return html
		}
	} else {
		log.Printf("DEBUG: Found products loop with ELSE. Body Length: %d, Else Length: %d\n", len(matches[1]), len(matches[2]))
	}

	productTemplate := matches[1]
	emptyTemplate := ""
	if len(matches) >= 3 {
		emptyTemplate = matches[2]
	}

	if len(products) == 0 {
		return re.ReplaceAllString(html, emptyTemplate)
	}

	var productsHTML strings.Builder
	for _, p := range products {
		item := productTemplate

		// Replace this.* with product data
		item = strings.ReplaceAll(item, "{{this.id}}", p.ID.String())
        item = strings.ReplaceAll(item, "{{this.slug}}", p.Slug) // Expose Slug
		item = strings.ReplaceAll(item, "{{this.name}}", p.Name)
		item = strings.ReplaceAll(item, "{{this.name}}", p.Name)
		
		// Strip HTML for list view and truncate
		cleanDesc := stripHTML(p.Description)
		if len(cleanDesc) > 120 {
			cleanDesc = cleanDesc[:120] + "..."
		}
		item = strings.ReplaceAll(item, "{{this.description}}", cleanDesc)

		item = strings.ReplaceAll(item, "{{this.image_url}}", p.ImageURL)

		// Debug Log to File
		f, _ := os.OpenFile("debug_log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if f != nil {
			fmt.Fprintf(f, "DEBUG: Processing Product %s (ID: %s)\n", p.Name, p.ID)
		}

		// Handle discount - MUST be done before generic price replacement
		hasDiscount := p.DiscountPrice != nil && *p.DiscountPrice > 0 && *p.DiscountPrice < p.Price
		if hasDiscount {
			item = strings.ReplaceAll(item, "{{this.original_price}}", formatPrice(p.Price))
			item = strings.ReplaceAll(item, "{{this.price}}", formatPrice(*p.DiscountPrice))
			discount := int((float64(p.Price-*p.DiscountPrice) / float64(p.Price)) * 100)
			item = strings.ReplaceAll(item, "{{this.discount_percent}}", strconv.Itoa(discount))
		} else {
			item = strings.ReplaceAll(item, "{{this.price}}", formatPrice(p.Price))
			// Clear discount placeholders if no discount
			item = strings.ReplaceAll(item, "{{this.original_price}}", "")
			item = strings.ReplaceAll(item, "{{this.discount_percent}}", "")
		}
		item = processConditional(item, "this.original_price", hasDiscount)
		item = processConditional(item, "this.discount_percent", hasDiscount) // Added support for discount badge conditional

		// Handle image
		item = processConditional(item, "this.image_url", p.ImageURL != "")

		// Replace parent reference
		item = strings.ReplaceAll(item, "{{../store.slug}}", store.Slug)
		item = strings.ReplaceAll(item, "{{../store.whatsapp}}", store.Whatsapp)

        // Generate Product Link
        link := ""
        targetSlug := p.ID.String()
        if p.Slug != "" {
            targetSlug = p.Slug
        }

        if isSubdomain {
             link = fmt.Sprintf("/product/%s", targetSlug)
        } else {
             link = fmt.Sprintf("/s/%s/p/%s", store.Slug, targetSlug)
        }
        item = strings.ReplaceAll(item, "{{product_link}}", link) // Helper for themes
        // Also support legacy manually constructed links if possible?
        // Themes usually construct links manually like /s/{{store.slug}}/p/{{id}}. 
        // We SHOULD override standard links if possible or provide a cleaner way.
        // NOTE: Direct string replacement of hardcoded links in template is hard.
        // BEST PRACTICE: Themes should use {{product_link}} 
        
        // But for existing structure, we might need a hack if themes use hardcoded structure.
        // Assuming themes use {{../store.slug}} and {{this.id}} or {{this.slug}}.
        // We can't easily change the structure defined in the template string itself without parsing.
        // HOWEVER, if we provide `store.slug` as empty string or something, it might break.
        
        // Let's assume for now themes might need update OR we just provide data.
        // User asked for "https://sevencols.sello.id/product/slug".
        // If template has `<a href="/s/{{store.slug}}/p/{{id}}">`
        // We can't change that HTML structure easily from here without regex replacing the href.
        
        // Let's try to regex replace standard link patterns
        // Pattern: /s/{{../store.slug}}/p/{{this.id}} OR /s/{{../store.slug}}/p/{{this.slug}}
        // If isSubdomain is true, replace that whole pattern with /product/{{this.slug}}
        
        // Since we are iterating, we can replace the output item's HREF if it matches known pattern.
        // But `item` is the template string.
        
        // Let's search/replace standard pattern in `item`:
        if isSubdomain {
             item = strings.ReplaceAll(item, fmt.Sprintf("/s/%s/p/{{this.id}}", store.Slug), fmt.Sprintf("/product/%s", targetSlug))
             item = strings.ReplaceAll(item, fmt.Sprintf("/s/%s/p/{{this.slug}}", store.Slug), fmt.Sprintf("/product/%s", targetSlug))
             // Also generic placeholder replacement
             item = strings.ReplaceAll(item, "/s/{{../store.slug}}/p/{{this.id}}", fmt.Sprintf("/product/%s", targetSlug))
             item = strings.ReplaceAll(item, "/s/{{../store.slug}}/p/{{this.slug}}", fmt.Sprintf("/product/%s", targetSlug))
        }

		// Expose Category ID for client-side filtering
		if p.CategoryID != nil {
			item = strings.ReplaceAll(item, "{{this.category_id}}", p.CategoryID.String())
		} else {
			item = strings.ReplaceAll(item, "{{this.category_id}}", "")
		}

		productsHTML.WriteString(item)
	}

	return re.ReplaceAllString(html, productsHTML.String())
}

func replaceCategoriesLoop(html string, categories []models.Category, store models.Store) string {
	re := regexp.MustCompile(`\{\{#each categories\}\}([\s\S]*?)\{\{/each\}\}`)
	matches := re.FindStringSubmatch(html)

	if len(matches) < 2 {
		return html
	}

	catTemplate := matches[1]

	var catsHTML strings.Builder
	for _, cat := range categories {
		item := catTemplate
		item = strings.ReplaceAll(item, "{{this.id}}", cat.ID.String())
		item = strings.ReplaceAll(item, "{{this.name}}", cat.Name)
		catsHTML.WriteString(item)
	}

	return re.ReplaceAllString(html, catsHTML.String())
}

func replaceProductImagesLoop(html string, product models.Product) string {
	// Ensure main image appears in thumbnails even if not in gallery DB
	allImages := make([]models.ProductImage, 0)
	if product.ImageURL != "" {
		alreadyInList := false
		for _, img := range product.Images {
			if img.ImageURL == product.ImageURL {
				alreadyInList = true
				break
			}
		}
		if !alreadyInList {
			allImages = append(allImages, models.ProductImage{ImageURL: product.ImageURL})
		}
	}
	allImages = append(allImages, product.Images...)

	// Process {{#if product.images}} conditional
	html = processConditional(html, "product.images", len(allImages) > 0)

	// Process {{#each product.images}} loop
	re := regexp.MustCompile(`(?s)\{\{#each product\.images\}\}([\s\S]*?)\{\{/each\}\}`)
	matches := re.FindStringSubmatch(html)
	if len(matches) < 2 {
		return html
	}

	imgTemplate := matches[1]
	var imgsHTML strings.Builder
	for i, img := range allImages {
		item := imgTemplate
		item = strings.ReplaceAll(item, "{{this.url}}", img.ImageURL)
		if i == 0 {
			item = strings.ReplaceAll(item, "{{#if @first}}active{{/if}}", "active")
		} else {
			item = strings.ReplaceAll(item, "{{#if @first}}active{{/if}}", "")
		}
		imgsHTML.WriteString(item)
	}

	return re.ReplaceAllString(html, imgsHTML.String())
}

func formatPrice(price int64) string {
	// Format number with thousand separator
	str := strconv.FormatInt(price, 10)
	n := len(str)
	if n <= 3 {
		return str
	}

	var result strings.Builder
	for i, c := range str {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteRune('.')
		}
		result.WriteRune(c)
	}
	return result.String()
}

// processConditional handles {{#if condition}}...{{else}}...{{/if}} logic manually with nesting support
func processConditional(html, condition string, keepTrue bool) string {
	f, _ := os.OpenFile("debug_renderer.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if f != nil {
		defer f.Close()
		fmt.Fprintf(f, "\n=== Processing Conditional: %s (KeepTrue: %v) ===\n", condition, keepTrue)
	}

	// 1. Find Opening Tag
	// Regex matches: {{ #if condition }} (tolerant to spaces)
	reOpenPattern := `(?i)\{\{\s*#if\s+` + regexp.QuoteMeta(condition) + `\s*\}\}`
	reOpen := regexp.MustCompile(reOpenPattern)
	
	// General regexes for nesting check
	reAnyIf := regexp.MustCompile(`(?i)\{\{\s*#if\s+[\s\S]+?\}\}`) // Matches any {{#if ...}}
	reClose := regexp.MustCompile(`\{\{\s*/if\s*\}\}`)
	reElse := regexp.MustCompile(`\{\{\s*else\s*\}\}`)

	for {
		// Find start of our target conditional
		loc := reOpen.FindStringIndex(html)
		if loc == nil {
			if f != nil { fmt.Fprintf(f, "No match found for opening tag\n") }
			break
		}
		
		startTagStart := loc[0]
		startTagEnd := loc[1]
		
		if f != nil { fmt.Fprintf(f, "Found open tag at [%d:%d]\n", startTagStart, startTagEnd) }
		
		// Find matching closing tag with nesting logic
		substring := html[startTagEnd:]
		depth := 1
		var closingTagStart, closingTagEnd int
		found := false
		
		// We scan the substring relative index
		currentIndex := 0
		
		for {
			// Find next {{#if}} or {{/if}}
			locNextIf := reAnyIf.FindStringIndex(substring[currentIndex:])
			locNextClose := reClose.FindStringIndex(substring[currentIndex:])
			
			// Determine which one comes first
			var nextKey string // "if" or "close"
			var nextLoc []int
			
			if locNextIf != nil && locNextClose != nil {
				if locNextIf[0] < locNextClose[0] {
					nextKey = "if"
					nextLoc = locNextIf
				} else {
					nextKey = "close"
					nextLoc = locNextClose
				}
			} else if locNextIf != nil {
				nextKey = "if"
				nextLoc = locNextIf
			} else if locNextClose != nil {
				nextKey = "close"
				nextLoc = locNextClose
			} else {
				// No more tags found, breaks loop (error case if depth > 0)
				break 
			}
			
			// Adjust loc to absolute index within substring
			absStart := currentIndex + nextLoc[0]
			absEnd := currentIndex + nextLoc[1]
			
			if nextKey == "if" {
				depth++
				if f != nil { fmt.Fprintf(f, "  Depth++ (%d) at rel pos %d (Found nested IF)\n", depth, absStart) }
			} else {
				depth--
				if f != nil { fmt.Fprintf(f, "  Depth-- (%d) at rel pos %d (Found /IF)\n", depth, absStart) }
			}
			
			currentIndex = absEnd
			
			if depth == 0 {
				closingTagStart = startTagEnd + absStart
				closingTagEnd = startTagEnd + absEnd
				found = true
				if f != nil { fmt.Fprintf(f, "Found matching close tag at [%d:%d]\n", closingTagStart, closingTagEnd) }
				break
			}
		}
		
		if !found {
			log.Printf("ERROR: No matching closing tag for %s (nested check failed)\n", condition)
			if f != nil { fmt.Fprintf(f, "ERROR: Not found matching closing tag\n") }
			return html
		}
		
		// 3. Extract Block Content
		blockContent := html[startTagEnd:closingTagStart]
		
		// 4. Handle Else (only top-level else!)
		// We need to find `{{else}}` that is at depth 1 (relative to this block)
		// Simpler approach: If we are keeping true, we take everything until the first TOP LEVEL else.
		// If keeping false, we take everything AFTER the first TOP LEVEL else.
		// But scanning for else requires the same nesting logic. To save complexity, assume else is rare/flat or handle simply.
		// Given we only use else in flat scenarios usually, or distinct blocks. 
		// But wait, the main products loop uses else.
		// Let's use a simpler heuristic for ELSE: split by {{else}}, but check if the else is inside a nested block.
		
		var finalContent string
		
		// Find valid ELSE at depth 0 relative to blockContent
		elsePos := -1
		depth = 0
		curIdx := 0
		for {
			locIf := reAnyIf.FindStringIndex(blockContent[curIdx:])
			locClose := reClose.FindStringIndex(blockContent[curIdx:])
			locE := reElse.FindStringIndex(blockContent[curIdx:])
			
			// Find nearest event
			minIdx := len(blockContent) + 1
			eventType := ""
			
			if locIf != nil && locIf[0] < minIdx { minIdx = locIf[0]; eventType = "if" }
			if locClose != nil && locClose[0] < minIdx { minIdx = locClose[0]; eventType = "close" }
			if locE != nil && locE[0] < minIdx { minIdx = locE[0]; eventType = "else" }
			
			if eventType == "" { break }
			
			realStart := curIdx + minIdx
			
			if eventType == "if" {
				depth++
				curIdx = realStart + (locIf[1] - locIf[0])
			} else if eventType == "close" {
				depth--
				curIdx = realStart + (locClose[1] - locClose[0])
			} else if eventType == "else" {
				if depth == 0 {
					elsePos = realStart
					break // Found our else
				}
				curIdx = realStart + (locE[1] - locE[0])
			}
		}
		
		if elsePos != -1 {
			if f != nil { fmt.Fprintf(f, "Found ELSE at relative pos %d\n", elsePos) }
			trueBlock := blockContent[:elsePos]
			// reElse match length
			elseMatch := reElse.FindString(blockContent[elsePos:])
			falseBlock := blockContent[elsePos+len(elseMatch):]
			
			if keepTrue {
				finalContent = trueBlock
			} else {
				finalContent = falseBlock
			}
		} else {
			if f != nil { fmt.Fprintf(f, "No ELSE found\n") }
			if keepTrue {
				finalContent = blockContent
			} else {
				finalContent = ""
			}
		}
		
		// 5. Reconstruct HTML
		html = html[:startTagStart] + finalContent + html[closingTagEnd:]
	}
	
	return html
}

// injectBranding adds the "Made with Sello" badge if required
func injectBranding(html string, store models.Store) string {
	showBranding := true
	var owner models.User
	if err := database.GetDB().Where("id = ?", store.UserID).Preload("Plan").First(&owner).Error; err == nil {
		if store.HideBranding && owner.Plan != nil && owner.Plan.RemoveBranding {
			showBranding = false
		}
	}

	if showBranding {
        // Updated badge HTML (simplified) - positioned above CTA
		badgeHTML := `<style>.sello-branding-badge{position:fixed;bottom:100px;left:50%;transform:translate(-50%) translateX(0);background:rgba(255,255,255,0.95);backdrop-filter:blur(8px);-webkit-backdrop-filter:blur(8px);padding:8px 16px;border-radius:9999px;display:flex;align-items:center;gap:8px;box-shadow:0 4px 12px rgba(0,0,0,0.08);font-family:"Inter",sans-serif;font-size:13px;color:#4b5563;text-decoration:none;transition:all 0.2s ease;z-index:9999;border:1px solid rgba(0,0,0,0.05)}.sello-branding-badge:hover{transform:translate(-50%) translateY(-2px);box-shadow:0 6px 16px rgba(0,0,0,0.12);color:#111827}.sello-branding-logo{height:16px;width:auto;display:block}@media (max-width:640px){.sello-branding-badge{bottom:90px;padding:6px 14px;font-size:12px}}</style><a href="https://sello.id" class="sello-branding-badge" target="_blank"><span>Dibuat dengan</span><img src="/static/images/sello-logo.webp" alt="Sello" class="sello-branding-logo"/></a>`
		if strings.Contains(html, "</body>") {
			return strings.Replace(html, "</body>", badgeHTML+"</body>", 1)
		}
		return html + badgeHTML
	}
	return html
}
