package middleware

import (
	"net/http"
	"sello/internal/database"
	"sello/internal/models"
	"sello/internal/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuthMiddleware - Protects routes using Local JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate using Local JWT Service
		claims, err := services.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set Context
		c.Set("user_id", claims.UserID)
		// We don't have email in claims unless we add it, but UserID is enough for GetMe
		
		c.Next()
	}
}

// RequireStoreMiddleware - For API routes, returns 403 if user has no stores
func RequireStoreMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		userID, err := uuid.Parse(userIDStr.(string))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		var count int64
		database.GetDB().Model(&models.Store{}).Where("user_id = ?", userID).Count(&count)
		
		if count == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"error":    "no_store",
				"message":  "You need to create a store first",
				"redirect": "/onboarding",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

