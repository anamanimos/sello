package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
)

func main() {
	// Change to project root
	wd, _ := os.Getwd()
	if filepath.Base(wd) == "seed" {
		os.Chdir(filepath.Join(wd, "..", ".."))
	}

	config.Load()
	database.Connect()
	database.GetDB().AutoMigrate(&models.Theme{})

	// Default config (Flat structure matching new theme.json)
	defaultConfig := models.ThemeConfig{
        // Colors
        "primary_color":    "#3b82f6",
        "secondary_color":  "#6366f1",
        "accent_color":     "#22c55e",
        "background_color": "#f9fafb",
        "text_color":       "#111827",
        
        // Texts
        "cta_button_text": "Hubungi Kami",
        "products_title":  "Produk Kami",
        "footer_text":     "© 2024 Toko Kami",
        "hero_title":      "Welcome",
        
        // Images
        "hero_image": "",
	}
	configJSON, _ := json.Marshal(defaultConfig)

	themes := []models.Theme{
		{
			Name:            "Starter",
			Slug:            "starter",
			Description:     "Tema default Sello - simple dan mudah dikustomisasi",
			ThumbnailURL:    "https://placehold.co/600x400/3b82f6/white?text=Starter",
			DirectoryPath:   "starter",
			DefaultConfig:   configJSON,
			IsActive:        true,
			IsPremium:       false,
			IsDefault:       true,
			SortOrder:       0,
		},
		{
			Name:            "Modern Blue",
			Slug:            "modern-blue",
			Description:     "Tema modern dengan warna biru yang profesional",
			ThumbnailURL:    "https://placehold.co/600x400/2563eb/white?text=Modern+Blue",
			DirectoryPath:   "modern-blue",
			DefaultConfig:   configJSON,
			IsActive:        true,
			IsPremium:       false,
			SortOrder:       1,
		},
		{
			Name:            "Elegant Purple",
			Slug:            "elegant-purple",
			Description:     "Tema elegan dengan nuansa ungu",
			ThumbnailURL:    "https://placehold.co/600x400/8b5cf6/white?text=Elegant+Purple",
			DirectoryPath:   "elegant-purple",
			DefaultConfig: func() []byte {
				cfg := models.ThemeConfig{
                    "primary_color":    "#8b5cf6",
                    "secondary_color":  "#a855f7",
                    "accent_color":     "#ec4899",
                    "background_color": "#faf5ff",
                    "text_color":       "#1e1b4b",
                    "cta_button_text": "Pesan Sekarang",
                    "products_title":  "Koleksi Eksklusif",
                    "footer_text":     "© 2024 Boutique",
				}
				j, _ := json.Marshal(cfg)
				return j
			}(),
			IsActive:  true,
			IsPremium: false,
			SortOrder: 2,
		},
		{
			Name:            "Fresh Green",
			Slug:            "fresh-green",
			Description:     "Tema segar dengan warna hijau natural",
			ThumbnailURL:    "https://placehold.co/600x400/10b981/white?text=Fresh+Green",
			DirectoryPath:   "fresh-green",
			DefaultConfig: func() []byte {
				cfg := models.ThemeConfig{
                    "primary_color":    "#10b981",
                    "secondary_color":  "#059669",
                    "accent_color":     "#f59e0b",
                    "background_color": "#ecfdf5",
                    "text_color":       "#064e3b",
                    "cta_button_text": "Pesan via WhatsApp",
                    "products_title": "Produk Alami",
                    "footer_text":    "© 2024 Nature Store",
				}
				j, _ := json.Marshal(cfg)
				return j
			}(),
			IsActive:  true,
			IsPremium: false,
			SortOrder: 3,
		},
	}

	for _, theme := range themes {
		var existing models.Theme
		if err := database.GetDB().Where("slug = ?", theme.Slug).First(&existing).Error; err != nil {
			database.GetDB().Create(&theme)
			log.Printf("✅ Created theme: %s", theme.Name)
		} else {
			// Update existing
			database.GetDB().Model(&existing).Updates(map[string]interface{}{
				"thumbnail_url":    theme.ThumbnailURL,
				"directory_path":   theme.DirectoryPath,
				"default_config":   theme.DefaultConfig,
				"is_default":       theme.IsDefault,
			})
			log.Printf("🔄 Updated theme: %s", theme.Name)
		}
	}

	log.Println("🎨 Theme seeding completed!")
}
