package handlers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sello/internal/database"
	"sello/internal/models"
)

// =========================================
// User Management (Admin Console)
// =========================================

// AdminListUsers - GET /console/api/users
func AdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := database.GetDB().Model(&models.User{})

	planID := c.Query("plan_id")
	status := c.Query("status") // active, suspended

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("email ILIKE ? OR name ILIKE ? OR phone ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if planID != "" {
		query = query.Where("plan_id = ?", planID)
	}

	if status == "suspended" {
		query = query.Where("is_suspended = true")
	} else if status == "active" {
		query = query.Where("is_suspended = false")
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var users []models.User
	// Preload Plan to display in table
	query.Preload("Plan").Order("created_at DESC").Limit(limit).Offset(offset).Find(&users)

	// Get store count for each user & LAZY BACKFILL PLAN
	type UserWithStoreCount struct {
		models.User
		StoreCount int64 `json:"store_count"`
	}
	
	result := make([]UserWithStoreCount, len(users))
	for i, user := range users {
		// Backfill Plan if missing (Lazy fix)
		if user.PlanID == nil {
			defaultPlan := uint(1)
			database.GetDB().Model(&user).Update("plan_id", defaultPlan)
			// Reload plan for response
			var p models.Plan
			database.GetDB().First(&p, defaultPlan)
			user.Plan = &p
			user.PlanID = &defaultPlan
		}

		var count int64
		database.GetDB().Model(&models.Store{}).Where("user_id = ?", user.ID).Count(&count)
		result[i] = UserWithStoreCount{User: user, StoreCount: count}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// AdminGetUser - GET /console/api/users/:id
func AdminGetUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.GetDB().Preload("Plan").First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Get user's stores
	var stores []models.Store
	database.GetDB().Where("user_id = ?", user.ID).Find(&stores)

	// Get user's subscriptions (invoices)
	var subscriptions []models.SubscriptionTransaction
	database.GetDB().Preload("Plan").Where("user_id = ?", user.ID).Order("created_at DESC").Find(&subscriptions)

	c.JSON(http.StatusOK, gin.H{
	})
}

// AdminCreateUserInput defines the payload for admin creating a user
type AdminCreateUserInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Whatsapp string `json:"whatsapp" binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
	PlanID   *uint  `json:"plan_id"`
}

// AdminCreateUser - POST /console/api/users
func AdminCreateUser(c *gin.Context) {
	var input AdminCreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if Email or Whatsapp exists
	var existingUser models.User
	if err := database.GetDB().Where("email = ? OR whatsapp = ?", input.Email, input.Whatsapp).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau WhatsApp sudah terdaftar"})
		return
	}

	// Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal enkripsi password"})
		return
	}

	// Handle basic phone prefix
	wa := input.Whatsapp
	if !strings.HasPrefix(wa, "62") && strings.HasPrefix(wa, "0") {
		wa = "62" + wa[1:]
	}

	// Handle Plan
	planID := uint(1) // Default to Free
	if input.PlanID != nil {
		planID = *input.PlanID
	}

	userID := uuid.New()
	newUser := models.User{
		ID:              userID,
		Email:           input.Email,
		Name:            input.Name,
		Whatsapp:        &wa,
		Password:        string(hashedPassword),
		PlanID:          &planID,
		IsPhoneVerified: true,  // Auto verify since created by admin
		IsEmailVerified: true,
	}

	if err := database.GetDB().Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat user"})
		return
	}

	logAdminAction(c, "create_user", "users", userID.String())

	// Reload with plan
	database.GetDB().Preload("Plan").First(&newUser, userID)

	c.JSON(http.StatusCreated, newUser)
}

// AdminSuspendUser - PUT /console/api/users/:id/suspend
func AdminSuspendUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.GetDB().First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsSuspended = true
	database.GetDB().Save(&user)

	// Log action
	logAdminAction(c, "suspend_user", "users", id)

	c.JSON(http.StatusOK, gin.H{"message": "User suspended"})
}

// AdminActivateUser - PUT /console/api/users/:id/activate
func AdminActivateUser(c *gin.Context) {
	id := c.Param("id")

	var user models.User
	if err := database.GetDB().First(&user, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.IsSuspended = false
	database.GetDB().Save(&user)

	logAdminAction(c, "activate_user", "users", id)

	c.JSON(http.StatusOK, gin.H{"message": "User activated"})
}

// AdminDeleteUser - DELETE /console/api/users/:id
func AdminDeleteUser(c *gin.Context) {
	id := c.Param("id")

	// Soft delete user and their stores
	if err := database.GetDB().Delete(&models.User{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	logAdminAction(c, "delete_user", "users", id)

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

// =========================================
// Store Management (Admin Console)
// =========================================

// AdminListStores - GET /console/api/stores
func AdminListStores(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status") // active, suspended, all

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := database.GetDB().Model(&models.Store{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ?", searchPattern, searchPattern)
	}

	if status == "suspended" {
		query = query.Where("is_suspended = true")
	} else if status == "active" {
		query = query.Where("is_suspended = false OR is_suspended IS NULL")
	}

	// Subquery to count products
	query.Select("stores.*, (SELECT COUNT(*) FROM products WHERE products.store_id = stores.id AND products.deleted_at IS NULL) as product_count")

	var total int64
	query.Model(&models.Store{}).Count(&total)

	offset := (page - 1) * limit
	var stores []models.Store
	query.Preload("Products").Order("created_at DESC").Limit(limit).Offset(offset).Find(&stores)

	c.JSON(http.StatusOK, gin.H{
		"data": stores,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// AdminGetStore - GET /console/api/stores/:id
func AdminGetStore(c *gin.Context) {
	id := c.Param("id")

	var store models.Store
	if err := database.GetDB().Preload("Products").First(&store, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	// Get owner info
	var owner models.User
	database.GetDB().First(&owner, "id = ?", store.UserID)

	// Get product count
	var productCount int64
	database.GetDB().Model(&models.Product{}).Where("store_id = ?", store.ID).Count(&productCount)

	c.JSON(http.StatusOK, gin.H{
		"store":         store,
		"owner":         owner,
		"product_count": productCount,
	})
}

// AdminVerifyStore - PUT /console/api/stores/:id/verify
func AdminVerifyStore(c *gin.Context) {
	id := c.Param("id")

	var store models.Store
	if err := database.GetDB().First(&store, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	store.IsVerified = true
	database.GetDB().Save(&store)

	logAdminAction(c, "verify_store", "stores", id)

	c.JSON(http.StatusOK, gin.H{"message": "Store verified"})
}

// AdminSuspendStore - PUT /console/api/stores/:id/suspend
func AdminSuspendStore(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Reason string `json:"reason" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reason is required"})
		return
	}

	var store models.Store
	if err := database.GetDB().First(&store, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	store.IsSuspended = true
	store.SuspensionReason = input.Reason
	database.GetDB().Save(&store)

	logAdminAction(c, "suspend_store", "stores", id)

	c.JSON(http.StatusOK, gin.H{"message": "Store suspended"})
}

// AdminActivateStore - PUT /console/api/stores/:id/activate
func AdminActivateStore(c *gin.Context) {
	id := c.Param("id")

	var store models.Store
	if err := database.GetDB().First(&store, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

	store.IsSuspended = false
	database.GetDB().Save(&store)

	logAdminAction(c, "activate_store", "stores", id)

	c.JSON(http.StatusOK, gin.H{"message": "Store activated"})
}

// AdminDeleteStore - DELETE /console/api/stores/:id
func AdminDeleteStore(c *gin.Context) {
	id := c.Param("id")

	if err := database.GetDB().Delete(&models.Store{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete store"})
		return
	}

	logAdminAction(c, "delete_store", "stores", id)

	c.JSON(http.StatusOK, gin.H{"message": "Store deleted"})
}

// AdminSeedStoreProducts - POST /console/api/stores/:id/seed
func AdminSeedStoreProducts(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Count int `json:"count"`
	}
	// Default to 10 if not specified or invalid
	if err := c.ShouldBindJSON(&input); err != nil || input.Count <= 0 {
		input.Count = 10
	}
    // Limit max to prevent abuse
    if input.Count > 50 {
        input.Count = 50
    }

	var store models.Store
	if err := database.GetDB().First(&store, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Store not found"})
		return
	}

    // Curated Realistic Products (Name, Description, ImageURL, BasePrice)
    type SeedProduct struct {
        Name        string
        Desc        string
        Image       string
        Price       int64
    }

    curatedList := []SeedProduct{
         {"Classic Leather Watch", "Elegan dan tahan lama. Jam tangan kulit asli ini cocok untuk acara formal maupun kasual. Water resistant 30m.", "https://images.unsplash.com/photo-1524592094714-0f0654e20314?auto=format&fit=crop&w=800&q=80", 1500000},
         {"Premium Knit Sweater", "Sweater rajut dengan bahan katun premium. Sangat nyaman dipakai di cuaca dingin namun tetap breathable.", "https://images.unsplash.com/photo-1576566588028-4147f3842f27?auto=format&fit=crop&w=800&q=80", 450000},
         {"Wireless Noise-Canceling Headphones", "Nikmati musik tanpa gangguan dengan teknologi noise-canceling aktif. Baterai tahan hingga 30 jam.", "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=800&q=80", 2800000},
         {"Minimalist Desk Lamp", "Lampu meja dengan desain minimalis modern. Dilengkapi fitur dimmable dan port USB untuk charging.", "https://images.unsplash.com/photo-1507473888900-52e1adad54cd?auto=format&fit=crop&w=800&q=80", 350000},
         {"Canvas Tote Bag", "Tas tote bag kanvas tebal dengan kompartemen laptop. Kuat, stylish, dan ramah lingkungan.", "https://images.unsplash.com/photo-1544816155-12df9643f363?auto=format&fit=crop&w=800&q=80", 125000},
         {"Vintage Film Camera", "Kamera analog 35mm yang direstorasi penuh. Sempurna untuk menangkap momen dengan estetika retro.", "https://images.unsplash.com/photo-1526170375885-4d8ecf77b99f?auto=format&fit=crop&w=800&q=80", 1850000},
         {"Organic Face Serum", "Serum wajah berbahan alami 100% organik. Mencerahkan dan melembabkan kulit tanpa bahan kimia keras.", "https://images.unsplash.com/photo-1620916566398-39f1143ab7be?auto=format&fit=crop&w=800&q=80", 290000},
         {"Ergonomic Office Chair", "Kursi kerja ergonomis dengan sandaran lumbal yang dapat disesuaikan. Dukung postur sehat seharian.", "https://images.unsplash.com/photo-1580480055273-228ff5388ef8?auto=format&fit=crop&w=800&q=80", 2200000},
         {"Handcrafted Ceramic Mug", "Cangkir keramik buatan tangan dengan glasir unik. Setiap cangkir memiliki karakter tersendiri.", "https://images.unsplash.com/photo-1514228742587-6b1558fcca3d?auto=format&fit=crop&w=800&q=80", 85000},
         {"Smart Home Speaker", "Speaker pintar dengan asisten suara terintegrasi. Kontrol rumah pintar Anda hanya dengan suara.", "https://images.unsplash.com/photo-1589492477829-5e65395b66cc?auto=format&fit=crop&w=800&q=80", 990000},
         {"Denim Jacket Classic", "Jaket denim potongan klasik yang tidak lekang oleh waktu. Bahan denim 14oz yang awet dan nyaman.", "https://images.unsplash.com/photo-1576871337632-b9aef4c17ab9?auto=format&fit=crop&w=800&q=80", 550000},
         {"Running Sneakers", "Sepatu lari ringan dengan bantalan responsif. Desain aerodinamis untuk performa maksimal.", "https://images.unsplash.com/photo-1542291026-7eec264c27ff?auto=format&fit=crop&w=800&q=80", 1100000},
         {"Succulent Plant Pot", "Pot tanaman sukulen mini dalam pot semen geometris. Dekorasi meja yang segar dan mudah dirawat.", "https://images.unsplash.com/photo-1485955900006-10f4d324d411?auto=format&fit=crop&w=800&q=80", 45000},
         {"Travel Backpack 40L", "Ransel travel 40L tahan air dengan banyak kompartemen. Teman terbaik untuk petualangan Anda.", "https://images.unsplash.com/photo-1553062407-98eeb64c6a62?auto=format&fit=crop&w=800&q=80", 750000},
         {"Sunglasses Aviator", "Kacamata hitam model aviator dengan lensa polarized. Perlindungan UV400 maksimal.", "https://images.unsplash.com/photo-1572635196237-14b3f281503f?auto=format&fit=crop&w=800&q=80", 180000},
         {"Mechanical Keyboard", "Keyboard mekanik dengan switch brown yang tactile namun tidak berisik. Full RGB backlight.", "https://images.unsplash.com/photo-1595225476474-87563907a212?auto=format&fit=crop&w=800&q=80", 1300000},
         {"Scented Soy Candle", "Lilin aromaterapi dari soy wax alami. Wangi Lavender & Vanilla yang menenangkan.", "https://images.unsplash.com/photo-1602825389660-1e5b8e9bea93?auto=format&fit=crop&w=800&q=80", 120000},
    }
    
	var products []models.Product
    var uploadErrors []string
    
	for i := 0; i < input.Count; i++ {
        // Pick from list cyclically
        seed := curatedList[i%len(curatedList)]
        
        // Add random number if count exceeds list to keep names unique
        suffix := ""
        if i >= len(curatedList) {
            suffix = fmt.Sprintf(" %d", (i/len(curatedList))+1)
        }
        
        name := seed.Name + suffix
        // Unique slug generation (add timestamp)
        slug := fmt.Sprintf("%s-%s-%d-%d", store.Slug, generateSlug(name), i+1, time.Now().UnixMilli())

        // Upload image to Cloudinary
        cloudinaryURL := ""
        if seed.Image != "" {
            url, _, err := UploadImageFromURL(seed.Image, store.Slug, slug)
            if err != nil {
                uploadErrors = append(uploadErrors, fmt.Sprintf("%s: %v", name, err))
            } else {
                cloudinaryURL = url
            }
        }

		products = append(products, models.Product{
			StoreID:     store.ID,
			Name:        name,
            Slug:        slug,
			Description: seed.Desc,
			Price:       seed.Price,
            ImageURL:    cloudinaryURL,
		})
	}

	if err := database.GetDB().Create(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to seed products: %v", err)})
		return
	}

	logAdminAction(c, "seed_products", "stores", id)

    response := gin.H{"message": fmt.Sprintf("Successfully injected %d products", len(products))}
    if len(uploadErrors) > 0 {
        response["image_warnings"] = uploadErrors
    }
	c.JSON(http.StatusOK, response)
}

// =========================================
// Platform Analytics (Admin Console)
// =========================================

// AdminGetAnalyticsOverview - GET /console/api/analytics/overview
func AdminGetAnalyticsOverview(c *gin.Context) {
	var totalUsers int64
	database.GetDB().Model(&models.User{}).Count(&totalUsers)

	var totalStores int64
	database.GetDB().Model(&models.Store{}).Count(&totalStores)

	var totalProducts int64
	database.GetDB().Model(&models.Product{}).Count(&totalProducts)

	var totalVisits int64
	database.GetDB().Model(&models.AnalyticsLog{}).Where("event_type = 'view_store' OR event_type = 'view_product'").Count(&totalVisits)

	var totalClicks int64
	database.GetDB().Model(&models.AnalyticsLog{}).Where("event_type = 'click_whatsapp'").Count(&totalClicks)

	c.JSON(http.StatusOK, gin.H{
		"users":    totalUsers,
		"stores":   totalStores,
		"products": totalProducts,
		"visits":   totalVisits,
		"clicks":   totalClicks,
	})
}

// Helper to log admin actions
func logAdminAction(c *gin.Context, action, resource, resourceID string) {
	adminID, _ := c.Get("admin_id")
	// In future: save to admin_audit_logs table
	_ = adminID
	_ = action
	_ = resource
	_ = resourceID
}

// =========================================
// Plan Management (Admin Console)
// =========================================

// AdminListPlans - GET /console/api/plans
func AdminListPlans(c *gin.Context) {
	var plans []models.Plan
	if err := database.GetDB().Order("price_monthly ASC").Find(&plans).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch plans: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, plans)
}

// AdminCreatePlan - POST /console/api/plans
func AdminCreatePlan(c *gin.Context) {
	var plan models.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.GetDB().Create(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create plan"})
		return
	}

	c.JSON(http.StatusCreated, plan)
}

// AdminUpdatePlan - PUT /console/api/plans/:id
func AdminUpdatePlan(c *gin.Context) {
	id := c.Param("id")
	var plan models.Plan
	if err := database.GetDB().First(&plan, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Plan not found"})
		return
	}

	var input models.Plan
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	plan.Name = input.Name
	plan.MaxStores = input.MaxStores
	plan.MaxProductsPerStore = input.MaxProductsPerStore
	plan.MaxStorageMB = input.MaxStorageMB
	plan.AllowCustomDomain = input.AllowCustomDomain
	plan.AccessPremiumThemes = input.AccessPremiumThemes
	plan.PriceMonthly = input.PriceMonthly
	plan.PriceYearly = input.PriceYearly
	plan.AllowCustomSlug = input.AllowCustomSlug
	plan.RemoveBranding = input.RemoveBranding
    plan.IsActive = input.IsActive

	if err := database.GetDB().Save(&plan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update plan"})
		return
	}

	c.JSON(http.StatusOK, plan)
}

// AdminDeletePlan - DELETE /console/api/plans/:id
func AdminDeletePlan(c *gin.Context) {
	id := c.Param("id")
	if err := database.GetDB().Delete(&models.Plan{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete plan"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Plan deleted"})
}

// =========================================
// Order Management (Admin Console)
// =========================================

// AdminListOrders - GET /console/api/orders
func AdminListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	search := c.Query("search")
	status := c.Query("status")
    storeID := c.Query("store_id")

	if page < 1 { page = 1 }
	if limit < 1 || limit > 100 { limit = 20 }

	query := database.GetDB().Model(&models.Order{})

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("customer_name ILIKE ? OR customer_email ILIKE ? OR id::text ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}
    
    if storeID != "" {
        query = query.Where("store_id = ?", storeID)
    }

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var orders []models.Order
	query.Preload("Items").Preload("Store").Order("created_at DESC").Limit(limit).Offset(offset).Find(&orders)

	c.JSON(http.StatusOK, gin.H{
		"data": orders,
		"meta": gin.H{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": int(math.Ceil(float64(total) / float64(limit))),
		},
	})
}

// AdminUpdateOrder - PUT /console/api/orders/:id/status
func AdminUpdateOrder(c *gin.Context) {
    id := c.Param("id")
    var input struct {
        Status string `json:"status" binding:"required"`
    }
    if err := c.ShouldBindJSON(&input); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    var order models.Order
    if err := database.GetDB().First(&order, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
        return
    }
    
    order.Status = models.OrderStatus(input.Status)
    database.GetDB().Save(&order)
    
    c.JSON(http.StatusOK, order)
}

