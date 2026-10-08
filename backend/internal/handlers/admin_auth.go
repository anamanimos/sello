package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
)

// AdminLogin - POST /console/api/auth/login
func AdminLogin(c *gin.Context) {
	var input models.AdminLoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var admin models.Admin
	if err := database.GetDB().Where("email = ?", input.Email).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !admin.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	now := time.Now()
	admin.LastLogin = &now
	database.GetDB().Save(&admin)

	// Generate JWT with admin prefix to distinguish from user tokens
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID.String(),
		"role":     admin.Role,
		"type":     "admin",
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": tokenString,
		"admin": admin,
	})
}

// AdminLogout - POST /console/api/auth/logout
func AdminLogout(c *gin.Context) {
	// Client-side token removal
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

// GetAdminMe - GET /console/api/auth/me
func GetAdminMe(c *gin.Context) {
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var admin models.Admin
	if err := database.GetDB().First(&admin, "id = ?", adminID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	c.JSON(http.StatusOK, admin)
}

// =========================================
// Admin Management (Super Admin Only)
// =========================================

// ListAdmins - GET /console/api/admins
func ListAdmins(c *gin.Context) {
	var admins []models.Admin
	database.GetDB().Order("created_at DESC").Find(&admins)
	c.JSON(http.StatusOK, gin.H{"data": admins})
}

// CreateAdmin - POST /console/api/admins
func CreateAdmin(c *gin.Context) {
	var input models.CreateAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if email already exists
	var existing models.Admin
	if err := database.GetDB().Where("email = ?", input.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	admin := models.Admin{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
		Role:     input.Role,
		IsActive: true,
	}

	if err := database.GetDB().Create(&admin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin"})
		return
	}

	c.JSON(http.StatusCreated, admin)
}

// UpdateAdmin - PUT /console/api/admins/:id
func UpdateAdmin(c *gin.Context) {
	id := c.Param("id")
	currentAdminID, _ := c.Get("admin_id")

	var admin models.Admin
	if err := database.GetDB().First(&admin, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
		return
	}

	// Prevent self-demotion from super_admin
	if admin.ID == currentAdminID.(uuid.UUID) && admin.Role == models.RoleSuperAdmin {
		// Check if trying to change own role
		var input models.UpdateAdminInput
		if err := c.ShouldBindJSON(&input); err == nil {
			if input.Role != "" && input.Role != models.RoleSuperAdmin {
				c.JSON(http.StatusForbidden, gin.H{"error": "Cannot demote yourself from super_admin"})
				return
			}
		}
	}

	var input models.UpdateAdminInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Email != "" {
		admin.Email = input.Email
	}
	if input.Name != "" {
		admin.Name = input.Name
	}
	if input.Role != "" {
		admin.Role = input.Role
	}
	if input.IsActive != nil {
		admin.IsActive = *input.IsActive
	}
	if input.Password != "" {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		admin.Password = string(hashedPassword)
	}

	database.GetDB().Save(&admin)
	c.JSON(http.StatusOK, admin)
}

// DeleteAdmin - DELETE /console/api/admins/:id
func DeleteAdmin(c *gin.Context) {
	id := c.Param("id")
	currentAdminID, _ := c.Get("admin_id")

	adminUUID, _ := uuid.Parse(id)
	if adminUUID == currentAdminID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete yourself"})
		return
	}

	if err := database.GetDB().Delete(&models.Admin{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete admin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Admin deleted"})
}
