package handlers

import (
	"fmt"
	"net/http"
	"sello/internal/database"
	"sello/internal/models"
	"sello/internal/services"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// Register input
type RegisterInput struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Whatsapp  string `json:"whatsapp" binding:"required"`
	Password  string `json:"password" binding:"required,min=8"`
	StoreName string `json:"store_name"`
}

// Login input
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register - POST /api/auth/register
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Check Registration Allowed Setting
	var allowRegSetting models.SystemSetting
	if err := database.GetDB().Where("key = ?", "allow_registration").First(&allowRegSetting).Error; err == nil {
		if allowRegSetting.Value == "false" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Registrasi sementara ditutup"})
			return
		}
	}

	// 2. Check Existing User (Email OR Whatsapp)
	var existingUser models.User
	if err := database.GetDB().Where("email = ? OR whatsapp = ?", input.Email, input.Whatsapp).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email atau WhatsApp sudah terdaftar"})
		return
	}

	// 3. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt password"})
		return
	}

	// 4. Create User
	userID := uuid.New()
	newUser := models.User{
		ID:              userID,
		Email:           input.Email,
		Name:            input.Name,
		Whatsapp:        &input.Whatsapp,
		Password:        string(hashedPassword),
		IsPhoneVerified: false,
		IsEmailVerified: false,
	}

	if err := database.GetDB().Create(&newUser).Error; err != nil {
		fmt.Println("[ERROR] Failed to save user:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// 5. Create Store (if provided)
	if input.StoreName != "" {
		slug := generateSlug(input.StoreName)
		
		// Check if slug already exists, append number if needed
		var existingCount int64
		database.GetDB().Model(&models.Store{}).Where("slug LIKE ?", slug+"%").Count(&existingCount)
		if existingCount > 0 {
			slug = fmt.Sprintf("%s-%d", slug, existingCount+1)
		}
		
		store := models.Store{
			UserID:   userID,
			Slug:     slug,
			Name:     input.StoreName,
			Whatsapp: input.Whatsapp,
			// Leave CustomDomain as empty - model should handle NULL
		}
		if err := database.GetDB().Create(&store).Error; err != nil {
			fmt.Printf("[ERROR] Failed to create store for user %s: %v\n", userID.String(), err)
		} else {
			fmt.Printf("[INFO] Store '%s' created successfully for user %s\n", slug, userID.String())
		}
	}

	// 6. Generate & Send OTP
	method := services.GetDefaultVerificationMethod() // whatsapp
	code, err := services.GenerateOTP(userID, method)
	if err != nil {
		// Log error but allow registration success? No, blocking.
		fmt.Println("[ERROR] Generate OTP Failed:", err)
	} else {
		// Send Async with error logging
		if method == "whatsapp" {
			go func() {
				if err := services.SendWhatsappOTP(input.Whatsapp, code); err != nil {
					fmt.Printf("[ERROR] WhatsApp OTP to %s failed: %v\n", input.Whatsapp, err)
				} else {
					fmt.Printf("[INFO] WhatsApp OTP sent to %s\n", input.Whatsapp)
				}
			}()
		} else {
			go func() {
				if err := services.SendEmailOTP(input.Email, code); err != nil {
					fmt.Printf("[ERROR] Email OTP to %s failed: %v\n", input.Email, err)
				} else {
					fmt.Printf("[INFO] Email OTP sent to %s\n", input.Email)
				}
			}()
		}
	}

	// Return Success
	c.JSON(http.StatusOK, gin.H{
		"message":             "Registration successful",
		"user_id":             userID.String(),
		"verification_needed": true,
	})
}

// Login - POST /api/auth/login
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Find User by Email
	var user models.User
	if err := database.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau Password salah"})
		return
	}

	// 2. Check Password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email atau Password salah"})
		return
	}

	// 3. Generate Local JWT
	token, err := services.GenerateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// 4. Return Response
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
		},
	})
}

// GetMe - GET /api/auth/me
func GetMe(c *gin.Context) {
	// Got from Middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}
	
	uidStr := userIDStr.(string)
	uid, _ := uuid.Parse(uidStr)

	// Is Beta Mode?
	var betaMode models.SystemSetting
	isBeta := false
	if err := database.GetDB().Where("key = ?", "beta_mode_enabled").First(&betaMode).Error; err == nil {
		if strings.TrimSpace(strings.ToLower(betaMode.Value)) == "true" {
			isBeta = true
		}
	}

	var user models.User
	// Preload Plan so we can see if user already has one
	if err := database.GetDB().Preload("Plan").First(&user, uid).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// SELF-HEALING: If no plan, assign default (Free)
	if user.PlanID == nil {
		defaultPlan := uint(1)
		database.GetDB().Model(&user).Update("plan_id", defaultPlan)
		// Reload
		var p models.Plan
		database.GetDB().First(&p, defaultPlan)
		user.Plan = &p
		user.PlanID = &defaultPlan
	}

	// INJECT BETA PLAN IF NEEDED
	if isBeta {
		// If user.Plan is nil or Free (1), override with Beta Plan
		if user.Plan == nil || user.Plan.ID == 1 {
			var defaultPlanIDSetting models.SystemSetting
			if err := database.GetDB().Where("key = ?", "beta_default_plan_id").First(&defaultPlanIDSetting).Error; err == nil {
				planID, _ := strconv.Atoi(defaultPlanIDSetting.Value)
				var betaPlan models.Plan
				if err := database.GetDB().First(&betaPlan, planID).Error; err == nil {
					user.Plan = &betaPlan
					user.PlanID = &betaPlan.ID
				}
			}
		}
	}

	// Get Stores
	var stores []models.Store
	database.GetDB().Where("user_id = ?", uid).Find(&stores)

	// Check if Google User (Empty password) - Simplistic check for now
	isGoogleUser := len(user.Password) == 0

	c.JSON(http.StatusOK, gin.H{
		"user":           user,
		"stores":         stores,
		"is_beta":        isBeta,
		"is_google_user": isGoogleUser,
	})
}

// Logout
func Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// Helper: generate slug from store name
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
	return slug
}

type VerificationInput struct {
	UserID string `json:"user_id" binding:"required"`
	Code   string `json:"code" binding:"required"`
	Type   string `json:"type" binding:"required"` 
}

// ActivateAccount
func ActivateAccount(c *gin.Context) {
	var input VerificationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uid, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID"})
		return
	}

	isValid := services.VerifyOTP(uid, input.Code, input.Type)
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode OTP Salah atau Kadaluarsa"})
		return
	}

	// Update Status
	updates := map[string]interface{}{}
	if input.Type == "whatsapp" {
		updates["is_phone_verified"] = true
	} else {
		updates["is_email_verified"] = true
	}
	database.GetDB().Model(&models.User{}).Where("id = ?", uid).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"message": "Akun berhasil diverifikasi!"})
}

// ResendOTP
func ResendOTP(c *gin.Context) {
	var input struct {
		UserID string `json:"user_id" binding:"required"`
		Type   string `json:"type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	uidPtr, err := uuid.Parse(input.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Logic: Get User -> Generate OTP -> Send
	// Need User details for Phone/Email
	var user models.User
	if err := database.GetDB().First(&user, "id = ?", uidPtr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code, err := services.GenerateOTP(uidPtr, input.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	if input.Type == "whatsapp" {
		// DEREFERENCE CHECK
		if user.Whatsapp != nil {
		    go services.SendWhatsappOTP(*user.Whatsapp, code)
		} else {
		    c.JSON(http.StatusBadRequest, gin.H{"error": "Nomor WhatsApp tidak ditemukan"})
		    return
		}
	} else {
		go services.SendEmailOTP(user.Email, code)
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

// CompleteOnboarding Input
type OnboardingInput struct {
	Whatsapp  string `json:"whatsapp" binding:"required"`
	StoreName string `json:"store_name" binding:"required"`
	Subdomain string `json:"subdomain"` // Optional, if empty use slug from name
}

// CompleteOnboarding - POST /api/auth/onboarding
func CompleteOnboarding(c *gin.Context) {
	// Got from Middleware
	uidStr := c.MustGet("user_id").(string)
	uid, _ := uuid.Parse(uidStr)

	var input OnboardingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Update User Whatsapp
	// Ensure format
	wa := input.Whatsapp
	// basic sanitization if frontend missed it
	if !strings.HasPrefix(wa, "62") && strings.HasPrefix(wa, "0") {
		wa = "62" + wa[1:]
	}
	
	if err := database.GetDB().Model(&models.User{}).Where("id = ?", uid).Update("whatsapp", wa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// 2. Create Store
	var slug string
	if input.Subdomain != "" {
		slug = strings.ToLower(input.Subdomain)
		// Sanitize just in case
        slug = strings.ReplaceAll(slug, " ", "-")
	} else {
		slug = generateSlug(input.StoreName)
	}

	// Check strict uniqueness for subdomain
	var count int64
	database.GetDB().Model(&models.Store{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Masyaallah, URL subdomain ini sudah dipakai orang lain. Coba yang lain ya!"})
		return
	}

	newStore := models.Store{
		UserID:   uid,
		Name:     input.StoreName,
		Slug:     slug,
		Whatsapp: wa,
	}

	if err := database.GetDB().Create(&newStore).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create store"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Onboarding completed",
		"store":   newStore,
	})
}

// UpdateProfile - PUT /api/auth/profile
func UpdateProfile(c *gin.Context) {
	uidStr := c.MustGet("user_id").(string)
	
	var input struct {
		Name     string `json:"name"`
		Whatsapp string `json:"whatsapp"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Whatsapp != "" {
		// Basic cleaning
		updates["whatsapp"] = input.Whatsapp
	}

	if err := database.GetDB().Model(&models.User{}).Where("id = ?", uidStr).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

// UpdatePassword - PUT /api/auth/password
func UpdatePassword(c *gin.Context) {
	uidStr := c.MustGet("user_id").(string)
	
	var input struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.GetDB().First(&user, "id = ?", uidStr).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Logic: If user has a password, verify OldPassword
	// If user has NO password (Google/Social), skip OldPassword check (allow setting password)
	hasPassword := len(user.Password) > 0

	if hasPassword {
		if input.OldPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Password lama diperlukan"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.OldPassword)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Password lama salah"})
			return
		}
	}

	// Encrypt New
	hashed, _ := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	
	if err := database.GetDB().Model(&user).Update("password", string(hashed)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah"})
}

// ForgotPassword - POST /api/auth/forgot-password
// Sends password reset link via WA or Email based on system setting
func ForgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email tidak valid"})
		return
	}

	// 1. Find User by Email
	var user models.User
	if err := database.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		// Don't reveal if user exists or not
		c.JSON(http.StatusOK, gin.H{"message": "Jika email terdaftar, Anda akan menerima link reset password"})
		return
	}

	// 2. Generate secure token
	token := uuid.New().String()
	
	// 3. Save token to database (expires in 24 hours)
	resetToken := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	
	// Delete old tokens for this user
	database.GetDB().Where("user_id = ?", user.ID).Delete(&models.PasswordResetToken{})
	
	if err := database.GetDB().Create(&resetToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat token reset"})
		return
	}

	// 4. Determine delivery method from system settings
	var methodSetting models.SystemSetting
	method := "email" // default
	if err := database.GetDB().Where("key = ?", "password_reset_method").First(&methodSetting).Error; err == nil {
		method = strings.ToLower(strings.TrimSpace(methodSetting.Value))
	}

	// 5. Build reset link
	resetLink := fmt.Sprintf("https://app.sello.id/reset-password?token=%s", token)

	// 6. Send via appropriate channel
	if method == "whatsapp" && user.Whatsapp != nil && *user.Whatsapp != "" {
		go services.SendPasswordResetWhatsApp(*user.Whatsapp, resetLink)
	} else {
		go services.SendPasswordResetEmail(user.Email, resetLink)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Jika email terdaftar, Anda akan menerima link reset password",
		"method":  method,
	})
}

// ResetPassword - POST /api/auth/reset-password
// Resets password using valid token
func ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid"})
		return
	}

	// 1. Find token
	var resetToken models.PasswordResetToken
	if err := database.GetDB().Where("token = ?", input.Token).First(&resetToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token tidak valid atau sudah kadaluarsa"})
		return
	}

	// 2. Check if token is expired
	if time.Now().After(resetToken.ExpiresAt) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token sudah kadaluarsa"})
		return
	}

	// 3. Check if token is already used
	if resetToken.UsedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Token sudah digunakan"})
		return
	}

	// 4. Hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	// 5. Update user password
	if err := database.GetDB().Model(&models.User{}).Where("id = ?", resetToken.UserID).Update("password", string(hashed)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui password"})
		return
	}

	// 6. Mark token as used
	now := time.Now()
	database.GetDB().Model(&resetToken).Update("used_at", now)

	c.JSON(http.StatusOK, gin.H{"message": "Password berhasil diubah! Silakan login dengan password baru."})
}
