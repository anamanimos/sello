package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
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

	fmt.Println("🌱 Seeding Users and Stores...")

    // Cleanup broken data (ID 0000...)
    db.Unscoped().Where("id = '00000000-0000-0000-0000-000000000000'").Delete(&models.User{})
    db.Unscoped().Where("email = 'budi@example.com'").Delete(&models.User{}) // Re-create fresh

	// 1. Create Dummy Users
	users := []models.User{
        {
			ID:       uuid.New(),
			Name:     "Anam Super",
			Email:    "cranam21@gmail.com", // Requested by user
		},
		{
			ID:       uuid.New(),
			Name:     "Budi Santoso",
			Email:    "budi.santoso@example.com", 
		},
		{
			ID:       uuid.New(),
			Name:     "Siti Aminah",
			Email:    "siti.aminah@example.com",
		},
		{
			ID:       uuid.New(),
			Name:     "Rudi Hartone",
			Email:    "rudi.hartone@example.com",
			IsSuspended: true,
		},
	}

	for i, u := range users {
		var existing models.User
		if err := db.Where("email = ?", u.Email).First(&existing).Error; err == nil {
			fmt.Printf("User %s already exists, skipping...\n", u.Email)
			users[i] = existing // Use existing for store creation
		} else {
			if err := db.Create(&users[i]).Error; err != nil {
				log.Printf("Failed to create user %s: %v\n", u.Email, err)
			} else {
				fmt.Printf("Created User: %s\n", u.Email)
			}
		}
	}

	// 2. Create Dummy Stores
	stores := []models.Store{
		{
			ID:          uuid.New(),
			UserID:      users[0].ID,
			Name:        "Budi Fashion",
			Slug:        "budifashion-99",
			Description: "Toko baju kekinian",
			IsVerified:  true,
			CustomDomain: "budifashion.local",
		},
		{
			ID:          uuid.New(),
			UserID:      users[1].ID,
			Name:        "Siti Cakes",
			Slug:        "siticakes-99",
			Description: "Kue basah dan kering",
			CustomDomain: "siticakes.local",
		},
		{
			ID:          uuid.New(),
			UserID:      users[2].ID,
			Name:        "Rudi Gadget",
			Slug:        "rudigadget-99",
			Description: "Jual hape murah",
			IsSuspended: true,
			CustomDomain: "rudigadget.local",
		},
	}

	for _, s := range stores {
		var existing models.Store
		if err := db.Where("slug = ?", s.Slug).First(&existing).Error; err == nil {
			fmt.Printf("Store %s already exists, skipping...\n", s.Slug)
		} else {
			if err := db.Create(&s).Error; err != nil {
				log.Printf("Failed to create store %s: %v\n", s.Slug, err)
			} else {
				fmt.Printf("Created Store: %s\n", s.Slug)
			}
		}
	}

    // 3. Create dummy products for analytics
    var createdStores []models.Store
    db.Find(&createdStores)
    
    for _, s := range createdStores {
        product := models.Product{
            StoreID: s.ID,
            Name: "Produk Contoh " + s.Name,
            Price: 100000,
            Description: "Deskripsi produk contoh",
        }
        
         if err := db.Where("name = ? AND store_id = ?", product.Name, s.ID).First(&models.Product{}).Error; err != nil {
             db.Create(&product)
             fmt.Printf("Created Product for %s\n", s.Name)
         }
         
         // Analytics logs
         db.Create(&models.AnalyticsLog{
             StoreID: s.ID,
             EventType: "view_store",
             PageURL: "/s/"+s.Slug,
             IPAddress: "127.0.0.1",
             UserAgent: "Mozilla/5.0",
             CreatedAt: time.Now(),
         })
    }

	fmt.Println("✅ Seeding completed!")
}
