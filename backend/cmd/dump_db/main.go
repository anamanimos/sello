package main

import (
	"fmt"
	"log"

	"sello/internal/database"
	"sello/internal/models"

	"github.com/joho/godotenv"
)

func main() {
    // Load .env from root
    if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error loading .env file")
	}

	database.Connect()

	var themes []models.Theme
	database.GetDB().Find(&themes)
	fmt.Println("--- THEMES ---")
	for _, t := range themes {
        fmt.Printf("ID: %s | Slug: %s | Dir: %s | Default: %v\n", t.ID, t.Slug, t.DirectoryPath, t.IsDefault)
	}

	var storeThemes []models.StoreTheme
	database.GetDB().Find(&storeThemes)
	fmt.Println("\n--- STORE THEMES ---")
	for _, st := range storeThemes {
        fmt.Printf("ID: %s | Store: %s | Theme: %s | Active: %v\nConfig: %s\n", st.ID, st.StoreID, st.ThemeID, st.IsActive, string(st.CustomConfig))
	}
}
