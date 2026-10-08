package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
)

// AdminAuthMiddleware verifies admin JWT token
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Check token type
		tokenType, _ := claims["type"].(string)
		if tokenType != "admin" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not an admin token"})
			c.Abort()
			return
		}

		adminIDStr, _ := claims["admin_id"].(string)
		adminID, err := uuid.Parse(adminIDStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid admin ID"})
			c.Abort()
			return
		}

		// Verify admin still exists and is active
		var admin models.Admin
		if err := database.GetDB().First(&admin, "id = ? AND is_active = true", adminID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Admin not found or inactive"})
			c.Abort()
			return
		}

		c.Set("admin_id", adminID)
		c.Set("admin_role", admin.Role)
		c.Set("admin", admin)
		c.Next()
	}
}

// RequireRole checks if admin has required role
func RequireRole(roles ...models.AdminRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, exists := c.Get("admin_role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Role not found"})
			c.Abort()
			return
		}

		role := adminRole.(models.AdminRole)
		
		// Super admin can do everything
		if role == models.RoleSuperAdmin {
			c.Next()
			return
		}

		// Check if role is in allowed roles
		for _, allowedRole := range roles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// RequireSuperAdmin shorthand for super admin only
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole(models.RoleSuperAdmin)
}
