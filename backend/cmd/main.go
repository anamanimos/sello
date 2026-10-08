package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"sello/config"
	"sello/internal/database"
	"sello/internal/handlers"
	"sello/internal/middleware"
	"sello/internal/models"
)

func main() {
	// Load configuration
	config.Load()
    // Sello Backend Initialized (Analytics Enabled)

	// Set Gin mode
	gin.SetMode(config.AppConfig.GinMode)

	// Connect to database
	database.Connect()

	// Auto migrate all models
	err := database.GetDB().AutoMigrate(
		&models.User{},
		&models.Store{},
		&models.Product{},
		&models.Category{},
		&models.Plan{},
		&models.OTP{},
		// &models.Theme{}, // Disk based, do not migrate
		&models.StoreTheme{},
		&models.SystemSetting{},
		&models.SubscriptionTransaction{},
		&models.ProductImage{},
	)
	if err != nil {
		log.Println("WARNING: AutoMigrate failed:", err)
	}

	// Manual migration for missing column (AutoMigrate failing on some envs)
	database.GetDB().Exec("ALTER TABLE themes ADD COLUMN IF NOT EXISTS directory_path VARCHAR(255)")
	// Manual migration for external_link
	database.GetDB().Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS external_link TEXT")
	// Manual migration for slug column (required for product URL routing)
	database.GetDB().Exec("ALTER TABLE products ADD COLUMN IF NOT EXISTS slug TEXT")
	// Manual migration for is_active (grandfathering)
	database.GetDB().Exec("ALTER TABLE plans ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT TRUE")
	// Manual migration for allow_external_link feature
	database.GetDB().Exec("ALTER TABLE plans ADD COLUMN IF NOT EXISTS allow_external_link BOOLEAN DEFAULT FALSE")
	// Manual migration for hide_branding feature
	database.GetDB().Exec("ALTER TABLE stores ADD COLUMN IF NOT EXISTS hide_branding BOOLEAN DEFAULT FALSE")

    // FORCE CREATE OTP TABLE (AutoMigrate flaky with Supabase Pooler)
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS otps (
            id SERIAL PRIMARY KEY,
            user_id UUID NOT NULL,
            code VARCHAR(10) NOT NULL,
            type VARCHAR(20),
            expires_at TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            used BOOLEAN DEFAULT FALSE
        );
        CREATE INDEX IF NOT EXISTS idx_otps_user_id ON otps(user_id);
    `)

    // FORCE CREATE PRODUCT_IMAGES TABLE (AutoMigrate may fail silently)
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS product_images (
            id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
            product_id UUID NOT NULL,
            image_url TEXT NOT NULL,
            image_size BIGINT DEFAULT 0,
            is_primary BOOLEAN DEFAULT FALSE,
            sort_order INT DEFAULT 0,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
        );
        CREATE INDEX IF NOT EXISTS idx_product_images_product_id ON product_images(product_id);
    `)

    // --- THEME MIGRATION (DB to Disk) ---
    // 1. Add theme_name column
    database.GetDB().Exec("ALTER TABLE store_themes ADD COLUMN IF NOT EXISTS theme_name VARCHAR(100)")
    
    // 2. Migrate data (Copy directory_path from themes table to store_themes.theme_name)
    // We check if themes table exists first to avoid error on subsequent runs
    var tableExists bool
    database.GetDB().Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'themes')").Scan(&tableExists)
    if tableExists {
        database.GetDB().Exec(`
            UPDATE store_themes 
            SET theme_name = (
                SELECT COALESCE(NULLIF(directory_path, ''), slug) 
                FROM themes 
                WHERE themes.id = store_themes.theme_id
            ) 
            WHERE theme_name IS NULL OR theme_name = ''
        `)
        
        // 3. Drop themes table (Cleanup)
        database.GetDB().Exec("DROP TABLE IF EXISTS themes CASCADE")
        
        // 4. Drop old foreign key column
        database.GetDB().Exec("ALTER TABLE store_themes DROP COLUMN IF EXISTS theme_id")
    }

    // Manual creation of system_settings table (AutoMigrate may fail silently)
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS system_settings (
            key VARCHAR(255) PRIMARY KEY,
            value TEXT,
            type VARCHAR(20),
            description TEXT,
            updated_at TIMESTAMP
        );
            INSERT INTO system_settings (key, value, type, description, updated_at)
            VALUES 
                ('verification_default', 'whatsapp', 'string', 'Default Verification Method', NOW()),
                ('default_register_plan_id', '1', 'number', 'Default Plan ID for New Users', NOW())
            ON CONFLICT (key) DO NOTHING;
        `)	
        // Ensure values exist if table already there (ON CONFLICT DO NOTHING above skips updates)
        database.GetDB().Exec("INSERT INTO system_settings (key, value, type, description, updated_at) VALUES ('default_register_plan_id', '1', 'number', 'Default Plan ID for New Users', NOW()) ON CONFLICT (key) DO NOTHING")

	// Seed logic moved to after AutoMigrate

    // Manual creation of subscription_transactions (AutoMigrate failing)
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS subscription_transactions (
            id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
            user_id UUID NOT NULL,
            plan_id BIGINT NOT NULL,
            amount DECIMAL(10,2) NOT NULL,
            duration_month INT DEFAULT 1,
            status VARCHAR(20) DEFAULT 'pending',
            payment_method VARCHAR(50),
            payment_proof TEXT,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );
        CREATE INDEX IF NOT EXISTS idx_subscription_transactions_user_id ON subscription_transactions(user_id);
        CREATE INDEX IF NOT EXISTS idx_subscription_transactions_status ON subscription_transactions(status);
    `)

    // Manual migration for coupon_id in subscription_transactions
    database.GetDB().Exec("ALTER TABLE subscription_transactions ADD COLUMN IF NOT EXISTS coupon_id BIGINT")

    // Manual creation of coupons table
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS coupons (
            id SERIAL PRIMARY KEY,
            store_id UUID,
            code VARCHAR(50) NOT NULL,
            type VARCHAR(20) DEFAULT 'fixed',
            value DECIMAL(10,2) NOT NULL,
            min_purchase DECIMAL(10,2) DEFAULT 0,
            max_discount DECIMAL(10,2) DEFAULT 0,
            usage_limit INT DEFAULT 0,
            used_count INT DEFAULT 0,
            expires_at TIMESTAMP WITH TIME ZONE,
            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            deleted_at TIMESTAMP WITH TIME ZONE
        );
        CREATE INDEX IF NOT EXISTS idx_coupons_store_id ON coupons(store_id);
        CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
    `)

	if err := database.GetDB().AutoMigrate(
		&models.Plan{},          // New (Must be before User)
		&models.User{},
		&models.Store{},
		&models.SystemSetting{}, 
		&models.OTP{},           
		&models.Product{},
		&models.Category{},
		&models.ProductImage{},
		// &models.Theme{}, // Removed (Disk Based)
		&models.StoreTheme{},
		&models.StoreThemeHistory{},
		&models.Coupon{},        // New
		&models.Order{},         // End-User Orders (Catalog Mode logs)
        &models.SubscriptionTransaction{}, // Merchant Plan Orders (SaaS)
		&models.OrderItem{},     
		&models.Admin{},
		&models.Template{},
		&models.AnalyticsLog{},
		&models.PasswordResetToken{}, // Password Reset
	); err != nil {
		log.Printf("Warning: Database AutoMigrate failed: %v", err)
	}
    
    // ==========================================
    // MIGRATION PATCHES (Run on every startup)
    // These ensure schema sync even if AutoMigrate fails
    // ==========================================
    log.Println("🔧 Running migration patches...")
    
    // Stores table patches
    database.GetDB().Exec(`ALTER TABLE stores ADD COLUMN IF NOT EXISTS is_suspended BOOLEAN DEFAULT FALSE`)
    database.GetDB().Exec(`ALTER TABLE stores ADD COLUMN IF NOT EXISTS suspension_reason TEXT`)
    
    // Products table patches
    database.GetDB().Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS image_url TEXT`)
    database.GetDB().Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS slug VARCHAR(255)`)
    database.GetDB().Exec(`ALTER TABLE products ADD COLUMN IF NOT EXISTS image_size BIGINT DEFAULT 0`)
    
    // Password Reset Tokens table (full create for safety)
    database.GetDB().Exec(`
        CREATE TABLE IF NOT EXISTS password_reset_tokens (
            id SERIAL PRIMARY KEY,
            user_id UUID NOT NULL,
            token VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            used_at TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    database.GetDB().Exec(`CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id)`)
    database.GetDB().Exec(`CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token)`)
    
    log.Println("✅ Migration patches completed")
    
    // Sync Themes from Disk - REMOVED (Disk Only Mode)
    // services.SyncThemesFromDisk()

    // Seed default plans if empty
    var planCount int64
    database.GetDB().Model(&models.Plan{}).Count(&planCount)
    if planCount == 0 {
        log.Println("🌱 Seeding default plans (Free, Pro, Business)...")
        plans := []models.Plan{
            {
                Name: "Free", MaxStores: 1, MaxProductsPerStore: 20, MaxStorageMB: 50,
                AllowCustomDomain: false, AllowCustomSlug: false, RemoveBranding: false, AccessPremiumThemes: false,
                PriceMonthly: 0, PriceYearly: 0,
            },
            {
                Name: "Pro", MaxStores: 5, MaxProductsPerStore: 100, MaxStorageMB: 500,
                AllowCustomDomain: true, AllowCustomSlug: true, RemoveBranding: true, AccessPremiumThemes: true,
                PriceMonthly: 49000, PriceYearly: 490000,
            },
             {
                Name: "Business", MaxStores: 20, MaxProductsPerStore: 1000, MaxStorageMB: 2048,
                AllowCustomDomain: true, AllowCustomSlug: true, RemoveBranding: true, AccessPremiumThemes: true,
                PriceMonthly: 149000, PriceYearly: 1490000,
            },
        }
        database.GetDB().Create(&plans)
    }

    // ==========================================
    // MIGRATION: BACKFILL PRODUCT SLUGS
    // ==========================================
    go func() {
        var products []models.Product
        // Find products with empty slug
        if err := database.GetDB().Where("slug IS NULL OR slug = ''").Find(&products).Error; err == nil && len(products) > 0 {
            log.Printf("🐌 Backfilling slugs for %d products...", len(products))
            for _, p := range products {
                newSlug := generateSlug(p.Name)
                // Ensure uniqueness
                originalSlug := newSlug
                counter := 1
                for {
                     var count int64
                     database.GetDB().Model(&models.Product{}).Where("store_id = ? AND slug = ? AND id != ?", p.StoreID, newSlug, p.ID).Count(&count)
                     if count == 0 {
                         break
                     }
                     newSlug = fmt.Sprintf("%s-%d", originalSlug, counter)
                     counter++
                }
                database.GetDB().Model(&p).Update("slug", newSlug)
            }
            log.Println("✅ Backfill slugs completed")
        }
    }()

    // ==========================================
    // SEED SUPER ADMIN (Must be after AutoMigrate)
    // ==========================================
	adminEmail := "admin@sello.id"
	adminHash := "$2a$10$CJKltHYXz75JGXQ0kXkgQu2cCHDazt.FMV5FWjWw1V1Rdx0juYgtm" // admin123
    log.Println("👤 Checking/Seeding Super Admin:", adminEmail)

    // Ensure Table Exists (Just in case AutoMigrate failed silently)
    database.GetDB().AutoMigrate(&models.Admin{})

    // Check if exists
    var admin models.Admin
	if err := database.GetDB().Unscoped().Where("email = ?", adminEmail).First(&admin).Error; err != nil {
        // Not found -> Create
		log.Println("👤 Creating Super Admin (Insert)...")
        database.GetDB().Exec("INSERT INTO admins (id, email, password, name, role, is_active, created_at, updated_at, deleted_at) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW(), NULL)",
            "00000000-0000-0000-0000-000000000001", adminEmail, adminHash, "Super Admin", "super_admin", true)
	} else {
        // Found -> Update Password & Active & Restore if deleted
		log.Println("👤 Updating Super Admin (Reset)...")
        database.GetDB().Exec("UPDATE admins SET password = ?, is_active = true, deleted_at = NULL WHERE email = ?", adminHash, adminEmail)
    }    

	// Get working directory (project root)
	wd, _ := os.Getwd()
	templatesDir := filepath.Join(wd, "web", "templates")
	staticPath := filepath.Join(wd, "web", "static")

	// Setup router
	r := gin.Default()

	// Load HTML templates (all templates in templates folder)
	r.LoadHTMLGlob(filepath.Join(templatesDir, "*.html"))

	// Serve static files
	r.Static("/static", staticPath)
	r.Static("/themes", "./web/themes")
	r.Static("/uploads", "./web/uploads")
	
	// Landing page for local development (accessible at /landing/)
	r.Static("/landing", "./landing")

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Domain Router middleware (detects subdomain/custom domain access)
	r.Use(middleware.DomainRouter("sello.id"))
	r.Use(middleware.StoreFromDomain())

	// ============================================
	// ADMIN CONSOLE PAGES (console.sello.id)
	// ============================================
	r.GET("/console", func(c *gin.Context) {
		c.Redirect(302, "/console/login")
	})
	r.GET("/console/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_login.html", nil)
	})
	r.GET("/console/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_dashboard.html", nil)
	})
	r.GET("/console/users", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_users.html", nil)
	})
	r.GET("/console/users/:id", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_user_detail.html", nil)
	})
	r.GET("/console/stores", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_stores.html", nil)
	})
	r.GET("/console/analytics", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_analytics.html", nil)
	})
	r.GET("/console/admins", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_admins.html", nil)
	})
	r.GET("/console/themes", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_themes.html", nil)
	})
	r.GET("/console/settings", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_settings_gateway.html", nil)
	})
	r.GET("/console/plans", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_plans.html", nil)
	})
	r.GET("/console/orders", func(c *gin.Context) {
		c.HTML(http.StatusOK, "console_orders.html", nil)
	})
    r.GET("/console/coupons", func(c *gin.Context) {
        c.HTML(http.StatusOK, "console_coupons.html", nil)
    })

	// ============================================
	// FRONTEND ROUTES (HTML Pages)
	// ============================================

	// Landing page is now handled by domainRoutes below (line ~153)
	// which checks access_type to serve store or landing page

	// Auth pages
	r.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})
	r.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", nil)
	})
	r.GET("/activate", func(c *gin.Context) {
		c.HTML(http.StatusOK, "activate_account.html", nil)
	})
	r.GET("/forgot-password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "forgot_password.html", nil)
	})
	r.GET("/reset-password", func(c *gin.Context) {
		c.HTML(http.StatusOK, "reset_password.html", nil)
	})
	r.GET("/auth/callback", func(c *gin.Context) {
		c.HTML(http.StatusOK, "auth_callback.html", nil)
	})
	r.GET("/onboarding", func(c *gin.Context) {
		c.HTML(http.StatusOK, "onboarding.html", nil)
	})

	// Dashboard pages
	r.GET("/dashboard", func(c *gin.Context) {
		c.HTML(http.StatusOK, "dashboard.html", gin.H{"Title": "Dashboard"})
	})
	r.GET("/dashboard/stores", func(c *gin.Context) {
		c.HTML(http.StatusOK, "stores.html", gin.H{"Title": "Toko"})
	})
	r.GET("/dashboard/products", func(c *gin.Context) {
		c.HTML(http.StatusOK, "products.html", gin.H{"Title": "Produk"})
	})
	r.GET("/dashboard/products/new", func(c *gin.Context) {
		c.HTML(http.StatusOK, "product-form.html", gin.H{"Title": "Tambah Produk"})
	})
	r.GET("/dashboard/products/:id/edit", func(c *gin.Context) {
		c.HTML(http.StatusOK, "product-edit.html", gin.H{"Title": "Edit Produk"})
	})
	r.GET("/dashboard/settings", func(c *gin.Context) {
		c.HTML(http.StatusOK, "settings.html", gin.H{"Title": "Pengaturan"})
	})
	r.GET("/dashboard/categories", func(c *gin.Context) {
		c.HTML(http.StatusOK, "categories.html", gin.H{"Title": "Kategori"})
	})
	r.GET("/dashboard/statistics", func(c *gin.Context) {
		c.HTML(http.StatusOK, "statistics.html", gin.H{"Title": "Statistik"})
	})
	r.GET("/dashboard/theme", func(c *gin.Context) {
		c.HTML(http.StatusOK, "theme.html", gin.H{"Title": "Tema"})
	})
	r.GET("/dashboard/theme/customize", func(c *gin.Context) {
		c.HTML(http.StatusOK, "theme_editor.html", nil)
	})
	r.GET("/dashboard/orders", func(c *gin.Context) {
		c.HTML(http.StatusOK, "orders.html", gin.H{"Title": "Pesanan"})
	})
	r.GET("/dashboard/coupons", func(c *gin.Context) {
		c.HTML(http.StatusOK, "coupons.html", gin.H{"Title": "Kupon"})
	})
	r.GET("/dashboard/profile", func(c *gin.Context) {
        c.HTML(http.StatusOK, "profile.html", gin.H{"Title": "Profil Saya"})
    })
	r.GET("/dashboard/billing", func(c *gin.Context) {
        // SERVER-SIDE BETA PROTECTION
        var betaMode models.SystemSetting
        // Check if beta_mode_enabled is "true" (robust check)
        if err := database.GetDB().Where("key = ?", "beta_mode_enabled").First(&betaMode).Error; err == nil {
            log.Printf("DEBUG: Beta Mode Check - Value: '%s'", betaMode.Value) // DEBUG LOG
            if strings.TrimSpace(strings.ToLower(betaMode.Value)) == "true" {
                c.Redirect(http.StatusFound, "/dashboard")
                return
            }
        } else {
             log.Printf("DEBUG: Beta Mode Check - Error or Not Found: %v", err)
        }
		c.HTML(http.StatusOK, "billing.html", gin.H{"Title": "Billing"})
	})
	r.GET("/dashboard/theme/preview/:slug", handlers.RenderThemePreviewPage)

	// Public store page (by slug) - path-based access (with theme support)
	r.GET("/s/:slug", handlers.RenderThemedStorePage)
	r.GET("/s/:slug/p/:id", handlers.RenderThemedProductDetail)
    
    

	// ============================================
	// SUBDOMAIN / CUSTOM DOMAIN ROUTES
	// ============================================
	// These routes handle access via:
	// - {slug}.sello.id (subdomain)
	// - custom-domain.com (verified custom domain)
	// 
	// In production, the DomainRouter middleware detects the host
	// and sets store_slug in context. These routes then serve the store.
	//
	// Note: For local development, use /s/:slug routes instead.
	// Subdomain routing requires proper DNS configuration.
	
	domainRoutes := r.Group("/")
	domainRoutes.Use(middleware.DomainRouter("sello.id"))
	{
		// When accessing via subdomain, "/" shows the store
		// When accessing via subdomain, "/" shows the store
		domainRoutes.GET("/", func(c *gin.Context) {
			accessType, _ := c.Get("access_type")

			if accessType == "console" {
				c.Redirect(http.StatusFound, "/console/login")
				return
			}

			if accessType == "subdomain" || accessType == "custom_domain" {
				handlers.RenderStoreFromDomain(c)
				return
			}
			// For main domain, redirect to login
			c.Redirect(http.StatusFound, "/login")
		})
		
		// Product detail on subdomain: /p/:id (not /s/:slug/p/:id)
		domainRoutes.GET("/p/:id", func(c *gin.Context) {
			accessType, _ := c.Get("access_type")
			if accessType == "subdomain" || accessType == "custom_domain" {
				handlers.RenderProductFromDomain(c)
				return
			}
			c.Redirect(302, "/")
		})

        // New Product Slug Route
        domainRoutes.GET("/product/:id", func(c *gin.Context) {
			accessType, _ := c.Get("access_type")
			if accessType == "subdomain" || accessType == "custom_domain" {
				handlers.RenderProductFromDomain(c)
				return
			}
			c.Redirect(302, "/")
		})
	}

	// ============================================
	// API ROUTES (JSON)
	// ============================================

	// Health check
	r.GET("/ping", handlers.Ping)
	r.GET("/health", handlers.HealthCheck)
	


	// Public	// Store Routes (Subdomain / Custom Domain)
	r.GET("/api/store/:slug/products", handlers.GetProductsByStore)
    
    // Public Plans
    r.GET("/api/plans", handlers.ListPlans)

	// Public analytics
	r.POST("/api/analytics/track", handlers.TrackClick)

	// List themes from disk (no database)
	r.GET("/api/themes-disk", handlers.ListThemes)
	r.GET("/api/themes/:id/preview", handlers.GetThemePreview)

	// E-commerce (Public Checkout & Payment)
	r.POST("/api/stores/:store_id/checkout", handlers.CreateOrder)
	r.GET("/api/orders/:order_id", handlers.GetOrder)
	r.POST("/api/orders/:order_id/pay", handlers.UploadPaymentProof)




	// ============================================
	// AUTH ROUTES
	// ============================================
	r.POST("/api/auth/register", handlers.Register)
	r.POST("/api/auth/login", handlers.Login)
	// Google OAuth
	r.GET("/api/auth/google/login", handlers.GoogleLogin)
	r.GET("/api/auth/google/callback", handlers.GoogleCallback)
	r.POST("/api/auth/logout", handlers.Logout)
	r.POST("/api/auth/activate", handlers.ActivateAccount)
	r.POST("/api/auth/resend-otp", handlers.ResendOTP)
	r.POST("/api/auth/forgot-password", handlers.ForgotPassword)
	r.POST("/api/auth/reset-password", handlers.ResetPassword)

	// Protected routes (require authentication)
	auth := r.Group("/api")
	auth.Use(middleware.AuthMiddleware())
	{
		// Auth
		auth.GET("/auth/me", handlers.GetMe)
		auth.POST("/auth/onboarding", handlers.CompleteOnboarding)
		auth.PUT("/auth/profile", handlers.UpdateProfile)
		auth.PUT("/auth/password", handlers.UpdatePassword)

		// Store management
		auth.GET("/stores", handlers.GetMyStores)
		auth.POST("/stores", handlers.CreateStore)
		auth.PUT("/stores/:store_id", handlers.UpdateStore)
		auth.DELETE("/stores/:store_id", handlers.DeleteStore)
		auth.GET("/stores/:store_id/stats", handlers.GetDashboardStats)
		
		// Custom Domain management
		auth.GET("/stores/:store_id/domain", handlers.GetDomainStatus)
		auth.PUT("/stores/:store_id/domain", handlers.SetCustomDomain)
		auth.POST("/stores/:store_id/domain/verify", handlers.VerifyCustomDomain)
		auth.DELETE("/stores/:store_id/domain", handlers.RemoveCustomDomain)

		// Product management
		auth.GET("/stores/:store_id/products", handlers.GetProductsByStoreID)
		auth.GET("/products/:id", handlers.GetProduct)
		auth.POST("/stores/:store_id/products", handlers.CreateProduct)
		auth.PUT("/products/:id", handlers.UpdateProduct)
		auth.DELETE("/products/:id", handlers.DeleteProduct)

		// Product images
		auth.GET("/products/:id/images", handlers.GetProductImages)
		auth.POST("/products/:id/images", handlers.AddProductImage)
		auth.PUT("/products/:id/images/:image_id/primary", handlers.SetPrimaryImage)
		auth.PUT("/products/:id/images/reorder", handlers.ReorderImages)
		auth.DELETE("/products/:id/images/:image_id", handlers.DeleteProductImage)

		// Upload & Storage
		auth.POST("/stores/:store_id/upload", handlers.UploadImage)
		auth.GET("/stores/:store_id/storage", handlers.GetStorageUsage)

		auth.DELETE("/images", handlers.DeleteImage)

		// Category management
		auth.GET("/stores/:store_id/categories", handlers.GetStoreCategories)
		auth.POST("/stores/:store_id/categories", handlers.CreateCategory)
		auth.PUT("/categories/:id", handlers.UpdateCategory)
		auth.DELETE("/categories/:id", handlers.DeleteCategory)


		// Order management
		auth.GET("/stores/:store_id/orders", handlers.GetStoreOrders)
		auth.PUT("/stores/:store_id/orders/:order_id/status", handlers.UpdateOrderStatus)

		// Coupon management
		auth.POST("/stores/:store_id/coupons", handlers.CreateCoupon)
		auth.GET("/stores/:store_id/coupons", handlers.GetCoupons)
		auth.DELETE("/coupons/:id", handlers.DeleteCoupon)

		// Theme management (user)
		auth.GET("/themes", handlers.ListThemes)

		auth.GET("/stores/:store_id/theme", handlers.GetStoreTheme)
		auth.PUT("/stores/:store_id/theme", handlers.SelectTheme)
		auth.POST("/stores/:store_id/theme/config", handlers.SaveThemeConfig)
		auth.GET("/stores/:store_id/theme/history", handlers.GetThemeHistory)
		auth.POST("/stores/:store_id/theme/restore", handlers.RestoreThemeHistory)

		// Billing (Merchant buys Plan)
		auth.GET("/billing/bank", handlers.GetBankInfo)
		auth.POST("/billing/subscribe", handlers.CreateSubscription)
		auth.POST("/billing/coupon/verify", handlers.VerifyCoupon) // Added this
		auth.POST("/billing/:id/proof", handlers.UploadSubscriptionProof)
	}

	// ============================================
	// ADMIN CONSOLE ROUTES (console.sello.id)
	// ============================================
	console := r.Group("/console/api")
	{
		// Public: Admin login
		console.POST("/auth/login", handlers.AdminLogin)
		console.POST("/auth/logout", handlers.AdminLogout)
		
		// Public: Email template preview (no auth needed - static templates)
		console.GET("/settings/email-preview", handlers.PreviewEmailTemplate)

		// Protected admin routes
		adminAuth := console.Group("")
		adminAuth.Use(middleware.AdminAuthMiddleware())
		{
			adminAuth.GET("/auth/me", handlers.GetAdminMe)

			// User Management
			adminAuth.GET("/users", handlers.AdminListUsers)
			adminAuth.POST("/users", middleware.RequireRole(models.RoleAdmin, models.RoleModerator), handlers.AdminCreateUser)
			adminAuth.GET("/users/:id", handlers.AdminGetUser)
			adminAuth.PUT("/users/:id/suspend", middleware.RequireRole(models.RoleAdmin, models.RoleModerator), handlers.AdminSuspendUser)
			adminAuth.PUT("/users/:id/activate", middleware.RequireRole(models.RoleAdmin), handlers.AdminActivateUser)
			adminAuth.DELETE("/users/:id", middleware.RequireSuperAdmin(), handlers.AdminDeleteUser)

			// Store Management
			adminAuth.GET("/stores", handlers.AdminListStores)
			adminAuth.GET("/stores/:id", handlers.AdminGetStore)
			adminAuth.PUT("/stores/:id/verify", middleware.RequireRole(models.RoleAdmin), handlers.AdminVerifyStore)
			adminAuth.PUT("/stores/:id/suspend", middleware.RequireRole(models.RoleAdmin, models.RoleModerator), handlers.AdminSuspendStore)
			adminAuth.PUT("/stores/:id/activate", middleware.RequireRole(models.RoleAdmin), handlers.AdminActivateStore)
			adminAuth.POST("/stores/:id/seed", middleware.RequireSuperAdmin(), handlers.AdminSeedStoreProducts) // New Seeder Route
			adminAuth.DELETE("/stores/:id", middleware.RequireSuperAdmin(), handlers.AdminDeleteStore)

			// Analytics
			adminAuth.GET("/analytics/overview", handlers.AdminGetAnalyticsOverview)

			// Admin Management (Super Admin only)
			adminAuth.GET("/admins", middleware.RequireSuperAdmin(), handlers.ListAdmins)
			adminAuth.POST("/admins", middleware.RequireSuperAdmin(), handlers.CreateAdmin)
			adminAuth.PUT("/admins/:id", middleware.RequireSuperAdmin(), handlers.UpdateAdmin)
			adminAuth.DELETE("/admins/:id", middleware.RequireSuperAdmin(), handlers.DeleteAdmin)

			// Theme Management (Super Admin only)
			adminAuth.GET("/themes", middleware.RequireSuperAdmin(), handlers.ListThemesAdmin)
			adminAuth.GET("/themes/:id", middleware.RequireSuperAdmin(), handlers.GetThemeAdmin)
			adminAuth.POST("/themes", middleware.RequireSuperAdmin(), handlers.CreateTheme)
			adminAuth.PUT("/themes/:id", middleware.RequireSuperAdmin(), handlers.UpdateTheme)
			adminAuth.DELETE("/themes/:id", middleware.RequireSuperAdmin(), handlers.DeleteTheme)

			// System Settings (Super Admin only)
			adminAuth.GET("/settings/gateway", middleware.RequireSuperAdmin(), handlers.GetSettings)
			adminAuth.POST("/settings/gateway", middleware.RequireSuperAdmin(), handlers.UpdateSettings)
			adminAuth.POST("/settings/test", middleware.RequireSuperAdmin(), handlers.TestGateway)
            
			// Plan Management
			adminAuth.GET("/plans", handlers.AdminListPlans)
			adminAuth.POST("/plans", middleware.RequireSuperAdmin(), handlers.AdminCreatePlan)
			adminAuth.PUT("/plans/:id", middleware.RequireSuperAdmin(), handlers.AdminUpdatePlan)
			adminAuth.DELETE("/plans/:id", middleware.RequireSuperAdmin(), handlers.AdminDeletePlan)

			// Order Management
			adminAuth.GET("/orders", handlers.AdminListOrders)
			adminAuth.PUT("/orders/:id/status", handlers.AdminUpdateOrder)

			// Subscription Management (Billing)
			adminAuth.GET("/subscriptions", handlers.AdminListSubscriptions)
			adminAuth.POST("/subscriptions/manual", handlers.AdminCreateManualSubscription) // New Manual Route
			adminAuth.PUT("/subscriptions/:id/approve", handlers.AdminApproveSubscription)

            // Coupon Management
			adminAuth.GET("/coupons", handlers.AdminListCoupons)
			adminAuth.POST("/coupons", middleware.RequireSuperAdmin(), handlers.AdminCreateCoupon)
			adminAuth.DELETE("/coupons/:id", middleware.RequireSuperAdmin(), handlers.AdminDeleteCoupon)
		}
	}

	// Start server
	port := config.AppConfig.Port
	log.Printf("🚀 Sello API starting on port %s", port)
	log.Printf("📄 Frontend: http://localhost:%s", port)
	r.Run(":" + port)
}

// Helper: generate slug from store name (Duplicated from auth.go for main package use)
func generateSlug(name string) string {
	slug := ""
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			slug += string(c)
		} else if c >= 'A' && c <= 'Z' {
			slug += string(c + 32)
		} else if c == ' ' || c == '-' {
			if len(slug) > 0 && slug[len(slug)-1] != '-' {
				slug += "-"
			}
		}
	}
	if len(slug) > 0 && slug[len(slug)-1] == '-' {
		slug = slug[:len(slug)-1]
	}
    // Handle empty slug (e.g. symbols only)
    if slug == "" {
        slug = "product"
    }
	return slug
}

