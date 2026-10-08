package main

import (
	"fmt"
	"log"

	"golang.org/x/crypto/bcrypt"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
)

func main() {
	// Load config
	config.Load()

	// Connect to database
	database.Connect()

	// Migrate Admin table
	database.GetDB().AutoMigrate(&models.Admin{})

	// Check if super admin already exists
	var count int64
	database.GetDB().Model(&models.Admin{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)

	if count > 0 {
		log.Println("Super admin already exists!")
		return
	}

	// Create first super admin
	email := "admin@sello.id"
	password := "admin123" // Change this!
	name := "Super Admin"

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	admin := models.Admin{
		Email:    email,
		Password: string(hashedPassword),
		Name:     name,
		Role:     models.RoleSuperAdmin,
		IsActive: true,
	}

	if err := database.GetDB().Create(&admin).Error; err != nil {
		log.Fatal("Failed to create super admin:", err)
	}

	fmt.Println("✅ Super Admin created successfully!")
	fmt.Println("   Email:", email)
	fmt.Println("   Password:", password)
	fmt.Println("")
	fmt.Println("⚠️  IMPORTANT: Change the password after first login!")
}
