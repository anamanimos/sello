package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"sello/internal/models"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Note: .env file not found, using system environment variables")
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
        PrepareStmt: false,
    })
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	targetEmail := "cranam21@gmail.com"
	fmt.Printf("🔍 Checking for user: %s\n", targetEmail)

	var user models.User
	// Unscoped to allow finding soft-deleted users too
	result := db.Unscoped().Where("email = ?", targetEmail).First(&user)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			fmt.Println("❌ User NOT FOUND in database.")
		} else {
			fmt.Printf("💥 Error querying database: %v\n", result.Error)
		}
	} else {
		fmt.Println("✅ User FOUND!")
		fmt.Printf("- ID: %s\n", user.ID)
		fmt.Printf("- Name: %s\n", user.Name)
		fmt.Printf("- CreatedAt: %v\n", user.CreatedAt)
		if user.DeletedAt.Valid {
			fmt.Printf("⚠️  User is SOFT DELETED (DeletedAt: %v)\n", user.DeletedAt.Time)
		} else {
			fmt.Println("- Status: Active (Not deleted)")
		}
	}

	var count int64
	db.Model(&models.User{}).Count(&count)
	fmt.Printf("\n📊 Total Active Users in DB: %d\n", count)
    
    if count > 0 {
        var allUsers []models.User
        db.Find(&allUsers)
        fmt.Println("--- User List ---")
        for _, u := range allUsers {
            fmt.Printf("- %s (ID: %s)\n", u.Email, u.ID)
        }
    }
}
