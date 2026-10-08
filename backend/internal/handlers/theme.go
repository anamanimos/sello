package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// ============================================
// HELPER: DISK OPERATIONS
// ============================================

// loadThemeFromDisk reads theme.json from web/themes/<folderName>
func loadThemeFromDisk(folderName string) (*models.Theme, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	jsonPath := filepath.Join(wd, "web", "themes", folderName, "theme.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var theme models.Theme
	if err := json.Unmarshal(content, &theme); err != nil {
		return nil, err
	}

	// Ensure ID/DirectoryPath is set to folder name (source of truth)
	theme.ID = folderName
	theme.DirectoryPath = folderName
	
	// Ensure Slug is populated
	if theme.Slug == "" {
		theme.Slug = folderName
	}

    // IsHidden and IsPremium come from JSON unmarshal automatically
    
    // Default to Active for now as these are disk-based available themes
    if theme.DefaultConfig != nil {
        theme.IsActive = true
    }

	return &theme, nil
}

// listAllThemes scans web/themes
// includeHidden: if true, returns all themes including hidden ones (for Admin)
// userID: if provided, filters themes by allowed_user_ids (empty string = no filtering)
func listAllThemes(includeHidden bool, userID string) ([]models.Theme, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	themesDir := filepath.Join(wd, "web", "themes")
	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil, err
	}

	var themes []models.Theme

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		theme, err := loadThemeFromDisk(entry.Name())
		if err == nil {
            // Filter hidden themes unless requested
            if theme.IsHidden && !includeHidden {
                continue
            }
            
            // Filter by AllowedUserIDs if the theme has restrictions
            // Skip this filter for admin (includeHidden=true)
            if !includeHidden && len(theme.AllowedUserIDs) > 0 {
                if userID == "" {
                    // Theme has restrictions but no user ID provided - hide it
                    continue
                }
                allowed := false
                for _, id := range theme.AllowedUserIDs {
                    if id == userID {
                        allowed = true
                        break
                    }
                }
                if !allowed {
                    continue
                }
            }
            
			themes = append(themes, *theme)
		}
	}
	
	// Sort by Order then Name
	sort.Slice(themes, func(i, j int) bool {
		if themes[i].SortOrder != themes[j].SortOrder {
			return themes[i].SortOrder < themes[j].SortOrder
		}
		return themes[i].Name < themes[j].Name
	})

	return themes, nil
}


// ============================================
// PUBLIC / API ENDPOINTS
// ============================================

// ListThemes - GET /api/themes (Public/Store Users)
// Hides hidden/exclusive themes and filters by user ID
func ListThemes(c *gin.Context) {
	// Get user ID from auth context if available
	userID := ""
	if claims, exists := c.Get("user_claims"); exists {
		if userClaims, ok := claims.(map[string]interface{}); ok {
			if id, ok := userClaims["user_id"].(string); ok {
				userID = id
			}
		}
	}
	// Also check for user_id directly set by auth middleware
	if uid, exists := c.Get("user_id"); exists {
		if id, ok := uid.(string); ok {
			userID = id
		}
	}

	themes, err := listAllThemes(false, userID) // False = Hide hidden themes
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load themes"})
		return
	}
	c.JSON(http.StatusOK, themes)
}

// GetThemePreview - GET /api/themes/:id/preview
// :id is now the folder name
func GetThemePreview(c *gin.Context) {
	themeName := c.Param("id") // folder name
	templateName := c.Query("template") // e.g. "index" or "product"
	
	// Default to index
	if templateName == "" {
		templateName = "index"
	}

	// Load Meta to check mappings
	meta := readDiskThemeMeta(themeName)
	
	// Determine filename based on template alias
	filename := "home.html" // Default for index
	if templateName == "product" {
		filename = "product.html"
	}
	
	// Override if defined in theme.json mapping
	if meta.Templates != nil {
		if val, ok := meta.Templates[templateName]; ok {
			filename = val
		}
	}
	
	wd, _ := os.Getwd()
	path := filepath.Join(wd, "web", "themes", themeName, filename)
	
	content, err := os.ReadFile(path)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template file not found", "path": filename})
		return
	}
	
	// Return HTML directly so it can be viewed in browser/iframe
	c.JSON(http.StatusOK, gin.H{"html_template": string(content)})
}

// ============================================
// STORE THEME ENDPOINTS
// ============================================

// GetStoreTheme - GET /api/stores/:store_id/theme
func GetStoreTheme(c *gin.Context) {
	storeID := c.Param("store_id")

	var storeTheme models.StoreTheme
	// Find Active Store Theme
	// Note: We used to search by Active=true, but usually a store just has one entry or we update the same one?
	// The models.StoreTheme has IsActive. Let's find the active one.
	
	result := database.GetDB().Where("store_id = ? AND is_active = ?", storeID, true).First(&storeTheme)
	
	if result.Error != nil {
		// Self-healing: Create default theme
		defaultThemeName := "starter"
		storeTheme = models.StoreTheme{
			StoreID:   uuid.MustParse(storeID),
			ThemeName: defaultThemeName,
			IsActive:  true,
		}
		if err := database.GetDB().Create(&storeTheme).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign default theme"})
			return
		}
	}

	// Inflate with Disk Data
	diskTheme, err := loadThemeFromDisk(storeTheme.ThemeName)
    var meta models.ThemeMeta
    
	if err == nil {
		storeTheme.Theme = diskTheme
        // Read full Meta/Schema for Editor
        meta = readDiskThemeMeta(storeTheme.ThemeName)
	} else {
		// Fallback if file missing
		storeTheme.Theme = &models.Theme{
			ID:   storeTheme.ThemeName,
			Name: fmt.Sprintf("%s (Missing)", storeTheme.ThemeName),
		}
	}

    // Return extended response
	c.JSON(http.StatusOK, gin.H{
        "id": storeTheme.ID,
        "store_id": storeTheme.StoreID,
        "theme_name": storeTheme.ThemeName,
        "custom_config": storeTheme.CustomConfig,
        "is_active": storeTheme.IsActive,
        "theme": storeTheme.Theme,
        "meta": meta, // Includes Settings Schema
    })
}

// SelectTheme - PUT /api/stores/:store_id/theme
func SelectTheme(c *gin.Context) {
	storeID := c.Param("store_id")
	
	var input struct {
		ThemeID      string                 `json:"theme_id"`
		CustomConfig map[string]interface{} `json:"custom_config"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Verify theme exists on disk
	diskTheme, err := loadThemeFromDisk(input.ThemeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Theme not found on server"})
		return
	}

	tx := database.GetDB().Begin()

	// 1. Deactivate current theme
	if err := tx.Model(&models.StoreTheme{}).
		Where("store_id = ?", storeID).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deactivate current theme"})
		return
	}

	// 2. Check if this theme was already added to store (inactive)
	var existingStoreTheme models.StoreTheme
	result := tx.Where("store_id = ? AND theme_name = ?", storeID, input.ThemeID).First(&existingStoreTheme)

	// Prepare config bytes
	var configBytes []byte
	if input.CustomConfig != nil {
		configBytes, _ = json.Marshal(input.CustomConfig)
	} else {
		configBytes, _ = json.Marshal(diskTheme.DefaultConfig)
	}

	if result.RowsAffected > 0 {
		// Activate it and Update Config
		existingStoreTheme.IsActive = true
		if input.CustomConfig != nil {
			existingStoreTheme.CustomConfig = configBytes
		}
		
		if err := tx.Save(&existingStoreTheme).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to activate theme"})
			return
		}
	} else {
		// Create new StoreTheme entry
		newStoreTheme := models.StoreTheme{
			StoreID:      uuid.MustParse(storeID),
			ThemeName:    input.ThemeID,
			CustomConfig: configBytes,
			IsActive:     true,
		}
		
		if err := tx.Create(&newStoreTheme).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to select theme"})
			return
		}
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Theme updated successfully"})
}

// SaveThemeConfig - POST /api/stores/:store_id/theme/config
func SaveThemeConfig(c *gin.Context) {
	storeID := c.Param("store_id")
	
	// Parse raw JSON body as map
	var config map[string]interface{}
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	configBytes, _ := json.Marshal(config)

	// Update Active Theme
	if err := database.GetDB().Model(&models.StoreTheme{}).
		Where("store_id = ? AND is_active = ?", storeID, true).
		Update("custom_config", configBytes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save config"})
		return
	}
	
	// Create History (Optional - simplified for now)
	// We can implement history tracking here if needed
	
	c.JSON(http.StatusOK, gin.H{"message": "Configuration saved"})
}

// GetThemeHistory - GET /api/stores/:store_id/theme/history
func GetThemeHistory(c *gin.Context) {
	storeID := c.Param("store_id")

	var history []models.StoreThemeHistory
	// Join with StoreTheme to filter by StoreID
	if err := database.GetDB().Joins("StoreTheme").
		Where("StoreTheme.store_id = ?", storeID).
		Order("created_at DESC").
		Limit(10).
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// RestoreThemeHistory - POST /api/stores/:store_id/theme/restore
func RestoreThemeHistory(c *gin.Context) {
	// storeID := c.Param("store_id")
	var input struct {
		HistoryID string `json:"history_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var history models.StoreThemeHistory
	if err := database.GetDB().Preload("StoreTheme").First(&history, "id = ?", input.HistoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "History record not found"})
		return
	}

	// Update the actual StoreTheme config
	if err := database.GetDB().Model(&models.StoreTheme{}).
		Where("id = ?", history.StoreThemeID).
		Update("custom_config", history.CustomConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore theme"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Theme restored successfully"})
}

// ============================================
// ADMIN (Console) ENDPOINTS - Placeholder
// ============================================
// Admin endpoints using DB for Themes are deprecated/removed.
// If needed, Admin can list from disk using ListThemes logic.

func ListThemesAdmin(c *gin.Context) {
	themes, err := listAllThemes(true, "") // True = Show all (including hidden), empty userID = no restriction
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load themes"})
        return
    }
    c.JSON(http.StatusOK, themes)
}

func GetThemeAdmin(c *gin.Context) {
	// Not implemented: Edit theme.json via API not supported yet
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Theme editing via API not supported"})
}

// CreateTheme - POST /console/api/themes
func CreateTheme(c *gin.Context) {
    var input struct {
        Name          string `json:"name"`
        Slug          string `json:"slug"`
        Description   string `json:"description"`
        ThumbnailURL  string `json:"thumbnail_url"`
        HTMLContent   string `json:"html_content"` // Base layout/home
        DefaultConfig string `json:"default_config"` // JSON string
        SortOrder     int    `json:"sort_order"`
        IsPremium     bool   `json:"is_premium"`
        IsActive      bool   `json:"is_active"` // Mapped to !IsHidden
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // validate slug
    if input.Slug == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Slug is required"})
        return
    }

    wd, _ := os.Getwd()
    themeDir := filepath.Join(wd, "web", "themes", input.Slug)

    // Check if exists
    if _, err := os.Stat(themeDir); !os.IsNotExist(err) {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Theme directory already exists"})
        return
    }

    // Create directories
    if err := os.MkdirAll(filepath.Join(themeDir, "layouts"), 0755); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
        return
    }
    os.MkdirAll(filepath.Join(themeDir, "pages"), 0755)
    os.MkdirAll(filepath.Join(themeDir, "assets"), 0755)

    // Write HTML Template (Home)
    // For simplicity, we put the content in pages/home.html and layouts/base.html
    // Real themes need separation, but for this "Quick Editor", we'll dump it in home.html
    // and assume base.html is minimal.
    
    baseHTML := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{ .Store.Name }}</title>
    <!-- Injected Assets -->
</head>
<body>
    {{ template "content" . }}
</body>
</html>`
    
    os.WriteFile(filepath.Join(themeDir, "layouts", "base.html"), []byte(baseHTML), 0644)
    
    // If input HTML is empty, provide default
    if input.HTMLContent == "" {
        input.HTMLContent = "<h1>Welcome to {{ .Store.Name }}</h1>"
    }
    os.WriteFile(filepath.Join(themeDir, "pages", "home.html"), []byte(input.HTMLContent), 0644)
    // Create dummy product page
    os.WriteFile(filepath.Join(themeDir, "pages", "product.html"), []byte("<h1>{{ .Product.Name }}</h1><p>{{ .Product.Description }}</p>"), 0644)

    // Parse Default Config
    var configMap map[string]interface{}
    if input.DefaultConfig != "" {
        json.Unmarshal([]byte(input.DefaultConfig), &configMap)
    }

    // Create theme.json
    meta := models.ThemeMeta{
        Name:         input.Name,
        Slug:         input.Slug,
        Description:  input.Description,
        Version:      "1.0.0",
        ThumbnailURL: input.ThumbnailURL,
        SortOrder:    input.SortOrder,
        IsPremium:    input.IsPremium,
        IsHidden:     !input.IsActive, // Active=True means Hidden=False
        Templates: map[string]string{
            "home": "pages/home.html",
            "product_detail": "pages/product.html",
        },
        DefaultConfig: configMap,
    }

    metaBytes, _ := json.MarshalIndent(meta, "", "  ")
    if err := os.WriteFile(filepath.Join(themeDir, "theme.json"), metaBytes, 0644); err != nil {
         c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write theme.json"})
         return
    }

    c.JSON(http.StatusCreated, gin.H{"message": "Theme created successfully"})
}

func UpdateTheme(c *gin.Context) {
    id := c.Param("id") // Folder name
    var input struct {
        Name          string `json:"name"`
        Description   string `json:"description"`
        ThumbnailURL  string `json:"thumbnail_url"`
        HTMLContent   string `json:"html_content"` 
        DefaultConfig string `json:"default_config"` 
        SortOrder     int    `json:"sort_order"`
        IsPremium     bool   `json:"is_premium"`
        IsActive      bool   `json:"is_active"`
    }

    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    wd, _ := os.Getwd()
    themeDir := filepath.Join(wd, "web", "themes", id)

    // Read existing meta
    meta := readDiskThemeMeta(id)
    
    // Update fields
    meta.Name = input.Name
    meta.Description = input.Description
    meta.ThumbnailURL = input.ThumbnailURL
    meta.SortOrder = input.SortOrder
    meta.IsPremium = input.IsPremium
    meta.IsHidden = !input.IsActive
    
    // Update config if provided
    var configMap map[string]interface{}
    if input.DefaultConfig != "" {
        if err := json.Unmarshal([]byte(input.DefaultConfig), &configMap); err == nil {
             meta.DefaultConfig = configMap
        }
    }

    // Write back theme.json
    metaBytes, _ := json.MarshalIndent(meta, "", "  ")
    os.WriteFile(filepath.Join(themeDir, "theme.json"), metaBytes, 0644)

    // Update HTML (home.html)
    if input.HTMLContent != "" {
         os.WriteFile(filepath.Join(themeDir, "pages", "home.html"), []byte(input.HTMLContent), 0644)
    }

    c.JSON(http.StatusOK, gin.H{"message": "Theme updated successfully"})
}

func DeleteTheme(c *gin.Context) {
    id := c.Param("id")
    wd, _ := os.Getwd()
    themeDir := filepath.Join(wd, "web", "themes", id)
    
    // Safety check: Don't delete system roots
    if id == "" || id == "." || id == ".." {
         c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid theme ID"})
         return
    }

    if err := os.RemoveAll(themeDir); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete theme"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"message": "Theme deleted"})
}


// --- Helper for old dependencies if any ---
// Some files might call getStoreThemeAndConfig internally (e.g. renderer)
// This needs to be exported or available if referenced by theme_renderer.go

// We need to fix theme_renderer.go as well because it might call GetThemeByID
// readDiskThemeMeta reads theme.json from disk
func readDiskThemeMeta(themeDir string) models.ThemeMeta {
	if themeDir == "" {
		return models.ThemeMeta{}
	}
	
	wd, _ := os.Getwd()
	jsonPath := filepath.Join(wd, "web", "themes", themeDir, "theme.json")

	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return models.ThemeMeta{}
	}
	
	var meta models.ThemeMeta
	if err := json.Unmarshal(content, &meta); err != nil {
        return models.ThemeMeta{}
    }
	
	return meta
}

// readDiskThemeConfig reads theme.json from disk to get latest default config
func readDiskThemeConfig(themeDir string) models.ThemeConfig {
	if themeDir == "" {
		return models.ThemeConfig{}
	}
	
	wd, _ := os.Getwd()
	jsonPath := filepath.Join(wd, "web", "themes", themeDir, "theme.json")
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		return models.ThemeConfig{}
	}
	
	var meta struct {
		DefaultConfig models.ThemeConfig `json:"default_config"`
	}
	json.Unmarshal(content, &meta)
	
	return meta.DefaultConfig
}
