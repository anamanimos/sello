package main

import (
	"fmt"
	"log"

	"sello/config"
	"sello/internal/database"
)

func main() {
	// Load config
	config.Load()

	// Connect to database
	database.Connect()
	db := database.GetDB()

	fmt.Println("🚀 Starting demo store seeding...")

	// 1. Create demo user with minimal fields
	var userID string
	db.Raw(`SELECT id::text FROM users WHERE email = 'demo@sello.id' LIMIT 1`).Scan(&userID)
	
	if userID == "" {
		// Try to create user with only email and name (minimal)
		result := db.Exec(`
			INSERT INTO users (id, email, name, created_at, updated_at)
			VALUES (gen_random_uuid(), 'demo@sello.id', 'Demo User', NOW(), NOW())
		`)
		if result.Error != nil {
			log.Println("Warning creating user:", result.Error)
		}
		db.Raw(`SELECT id::text FROM users WHERE email = 'demo@sello.id' LIMIT 1`).Scan(&userID)
		if userID != "" {
			fmt.Println("✅ Demo user created")
		} else {
			log.Fatal("Failed to create demo user")
		}
	} else {
		fmt.Println("✅ Demo user exists:", userID)
	}

	// 2. Create demo store
	var storeID string
	db.Raw(`SELECT id::text FROM stores WHERE slug = 'demo' LIMIT 1`).Scan(&storeID)
	
	if storeID == "" {
		result := db.Exec(`
			INSERT INTO stores (id, user_id, slug, name, description, whatsapp, is_verified, created_at, updated_at)
			VALUES (
				gen_random_uuid(), 
				?::uuid,
				'demo',
				'Toko Demo Sello',
				'Ini adalah toko demo untuk menunjukkan fitur-fitur Sello. Buat toko seperti ini gratis!',
				'628123456789',
				true,
				NOW(),
				NOW()
			)
		`, userID)
		if result.Error != nil {
			log.Fatal("Failed to create store:", result.Error)
		}
		db.Raw(`SELECT id::text FROM stores WHERE slug = 'demo' LIMIT 1`).Scan(&storeID)
		fmt.Println("✅ Demo store created")
	} else {
		fmt.Println("✅ Demo store exists:", storeID)
	}
	fmt.Println("   Store ID:", storeID)

	// 3. Create store theme with starter template
	var themeID string
	db.Raw(`SELECT id::text FROM store_themes WHERE store_id = ?::uuid LIMIT 1`, storeID).Scan(&themeID)
	if themeID == "" {
		result := db.Exec(`
			INSERT INTO store_themes (id, store_id, theme_name, is_active, created_at, updated_at)
			VALUES (gen_random_uuid(), ?::uuid, 'starter', true, NOW(), NOW())
		`, storeID)
		if result.Error != nil {
			log.Println("Warning creating theme:", result.Error)
		} else {
			fmt.Println("✅ Demo theme created: starter")
		}
	} else {
		fmt.Println("✅ Demo theme exists")
	}

	// 4. Delete existing demo products
	db.Exec(`DELETE FROM products WHERE store_id = ?::uuid`, storeID)
	fmt.Println("🗑️  Cleared existing demo products")

	// 5. Create 10 demo products
	products := []struct {
		Name          string
		Price         int64
		DiscountPrice *int64
		Description   string
		ImageURL      string
	}{
		{"Headphone Bluetooth Premium", 450000, intPtr(399000), "Headphone wireless dengan noise cancelling aktif.", "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=400&h=400&fit=crop"},
		{"Jaket Denim Classic", 275000, nil, "Jaket denim berkualitas tinggi dengan potongan classic.", "https://images.unsplash.com/photo-1551028719-00167b16eac5?w=400&h=400&fit=crop"},
		{"Tas Ransel Minimalist", 185000, intPtr(165000), "Tas ransel minimalist dengan bahan anti air.", "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?w=400&h=400&fit=crop"},
		{"Smartwatch Sport Edition", 850000, intPtr(799000), "Smartwatch dengan heart rate monitor dan GPS.", "https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=400&h=400&fit=crop"},
		{"Lampu Meja LED Modern", 125000, nil, "Lampu meja dengan 3 tingkat kecerahan.", "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&h=400&fit=crop"},
		{"Sepatu Sneakers Urban", 320000, intPtr(289000), "Sneakers nyaman untuk aktivitas sehari-hari.", "https://images.unsplash.com/photo-1542291026-7eec264c27ff?w=400&h=400&fit=crop"},
		{"Power Bank 20000mAh", 199000, nil, "Power bank dengan fast charging 22.5W.", "https://images.unsplash.com/photo-1609091839311-d5365f9ff1c5?w=400&h=400&fit=crop"},
		{"Kacamata Hitam Polarized", 145000, intPtr(120000), "Kacamata dengan lensa polarized UV protection.", "https://images.unsplash.com/photo-1572635196237-14b3f281503f?w=400&h=400&fit=crop"},
		{"Speaker Bluetooth Portable", 225000, nil, "Speaker portable waterproof IPX7.", "https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?w=400&h=400&fit=crop"},
		{"Topi Baseball Cap", 85000, intPtr(75000), "Topi baseball dengan bahan katun premium.", "https://images.unsplash.com/photo-1588850561407-ed78c282e89b?w=400&h=400&fit=crop"},
	}

	for i, p := range products {
		var result error
		if p.DiscountPrice != nil {
			result = db.Exec(`
				INSERT INTO products (id, store_id, name, price, discount_price, description, image_url, created_at, updated_at)
				VALUES (gen_random_uuid(), ?::uuid, ?, ?, ?, ?, ?, NOW(), NOW())
			`, storeID, p.Name, p.Price, *p.DiscountPrice, p.Description, p.ImageURL).Error
		} else {
			result = db.Exec(`
				INSERT INTO products (id, store_id, name, price, description, image_url, created_at, updated_at)
				VALUES (gen_random_uuid(), ?::uuid, ?, ?, ?, ?, NOW(), NOW())
			`, storeID, p.Name, p.Price, p.Description, p.ImageURL).Error
		}

		if result != nil {
			log.Printf("Warning creating product %d: %v", i+1, result)
		} else {
			fmt.Printf("✅ Product %d: %s\n", i+1, p.Name)
		}
	}

	fmt.Println("\n🎉 Demo store seeding complete!")
	fmt.Println("📍 Access at: /s/demo")
}

func intPtr(i int64) *int64 {
	return &i
}
