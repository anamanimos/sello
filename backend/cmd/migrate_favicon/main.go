package main

import (
	"log"
	"sello/config"
	"sello/internal/database"
)

func main() {
	config.Load()
	database.Connect()

	// Manually add the column
	err := database.GetDB().Exec("ALTER TABLE stores ADD COLUMN IF NOT EXISTS favicon_url TEXT").Error
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Migration successful: favicon_url column added.")
}
