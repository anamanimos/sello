package services

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"sello/internal/database"
	"sello/internal/models"
)

// HandleDowngrade enforces plan limits on a user's account
// It suspends extra stores and resets premium features if not allowed
func HandleDowngrade(userID uuid.UUID) error {
	db := database.GetDB()
	var user models.User
	if err := db.Preload("Plan").First(&user, userID).Error; err != nil {
		return err
	}

	// Default Limits (Free - fallback)
	maxStores := 1
	// allowCustomDomain := false // Handled by middleware
	accessPremium := false

	if user.Plan != nil {
		maxStores = user.Plan.MaxStores
		// allowCustomDomain = user.Plan.AllowCustomDomain
		accessPremium = user.Plan.AccessPremiumThemes
	}

	// 1. Handle Stores (Soft Lock)
	var stores []models.Store
	db.Where("user_id = ?", userID).Order("created_at asc").Find(&stores)

	for i, store := range stores {
		shouldSuspend := false
		
		// If index is greater than or equal to maxStores, it's an overflow store
		// (e.g. maxStores=1, index 0 is kept, index 1 is suspended)
		if i >= maxStores {
			shouldSuspend = true
		}

		// Only update if changed
		if shouldSuspend != store.IsSuspended {
			// If we are suspending, set IsSuspended = true
			// If we are unsuspending (upgrading), set IsSuspended = false
			store.IsSuspended = shouldSuspend
			db.Save(&store)
		}
		
		// 2. Handle Premium Themes
		if !accessPremium && !shouldSuspend {
			var storeTheme models.StoreTheme
			// Get Active Theme
			if db.Where("store_id = ? AND is_active = ?", store.ID, true).First(&storeTheme).Error == nil {
				// Check metadata if premium
				meta := ReadThemeMeta(storeTheme.ThemeName)
				if meta.IsPremium {
					// Downgrade to starter
					storeTheme.IsActive = false
					db.Save(&storeTheme)

					// Activate starter
					baseTheme := "starter"
					var starterTheme models.StoreTheme
					if db.Where("store_id = ? AND theme_name = ?", store.ID, baseTheme).First(&starterTheme).Error == nil {
						starterTheme.IsActive = true
						db.Save(&starterTheme)
					} else {
						// Create new entry
						db.Create(&models.StoreTheme{
							StoreID:   store.ID,
							ThemeName: baseTheme,
							IsActive:  true,
						})
					}
				}
			}
		}
	}

	return nil
}

// ReadThemeMeta reads theme.json from disk
func ReadThemeMeta(themeDir string) models.ThemeMeta {
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
