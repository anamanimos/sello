package main

import (
	"encoding/json"
	"fmt"
	"log"

	"sello/internal/database"
	"sello/internal/models"
	"sello/internal/services"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func main() {
    // Load .env
    if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

    // 0. Fix Schema Constraints (Legacy columns)
    fmt.Println("Fixing schema constraints...")
    database.GetDB().Exec("ALTER TABLE themes ALTER COLUMN html_template DROP NOT NULL")
    database.GetDB().Exec("ALTER TABLE themes ALTER COLUMN product_template DROP NOT NULL")

    // 1. Force Sync
    fmt.Println("Running SyncThemesFromDisk...")
    services.SyncThemesFromDisk()

    // 2. Verify Amerygo exists
    var amerygo models.Theme
    if err := database.GetDB().Where("slug = ?", "amerygo").First(&amerygo).Error; err != nil {
        log.Fatalf("Failed to find Amerygo after sync: %v", err)
    }
    fmt.Printf("Found Amerygo! ID: %s, Dir: %s\n", amerygo.ID, amerygo.DirectoryPath)

    // 3. Fix active store theme
    // Find the store (lazy way: use the ID from debug logs)
    storeID := "32165f56-dcfa-4953-8de8-d0179f297c45"
    
    // Deactivate all
    database.GetDB().Model(&models.StoreTheme{}).Where("store_id = ?", storeID).Update("is_active", false)

    // Check if store theme exists for Amerygo
    var storeTheme models.StoreTheme
    err := database.GetDB().Where("store_id = ? AND theme_id = ?", storeID, amerygo.ID).First(&storeTheme).Error

    if err != nil {
        fmt.Println("Creating new StoreTheme for Amerygo...")
        configBytes, _ := json.Marshal(map[string]interface{}{})
        storeTheme = models.StoreTheme{
            StoreID:      parseUUID(storeID),
            ThemeID:      amerygo.ID,
            IsActive:     true,
            CustomConfig: configBytes,
        }
        if err := database.GetDB().Create(&storeTheme).Error; err != nil {
             log.Fatalf("Failed to create store theme: %v", err)
        }
    } else {
        fmt.Println("Updating existing StoreTheme for Amerygo...")
        storeTheme.IsActive = true
        database.GetDB().Save(&storeTheme)
    }
    
    fmt.Println("Repair complete. Amerygo should be active now.")
}

func parseUUID(s string) uuid.UUID {
    id, _ := uuid.Parse(s)
    return id
}
