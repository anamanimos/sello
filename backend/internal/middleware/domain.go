package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"sello/internal/database"
	"sello/internal/models"
)

// DomainRouter detects whether the request is for:
// - Main site (sello.id, www.sello.id)
// - Admin app (app.sello.id)
// - Store subdomain ({slug}.sello.id)
// - Custom domain (www.tokoabc.com)
func DomainRouter(baseDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host

		// Remove port if present (for local development)
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Check for localhost/development
		if host == "localhost" || host == "127.0.0.1" {
			c.Set("access_type", "development")
			c.Next()
			return
		}

		// Check if it's the main domain or www
		if host == baseDomain || host == "www."+baseDomain {
			c.Set("access_type", "main")
			c.Next()
			return
		}

		// Check for app subdomain (admin dashboard)
		if host == "app."+baseDomain {
			c.Set("access_type", "admin")
			c.Next()
			return
		}

		// Check for console subdomain (superadmin console)
		if host == "console."+baseDomain {
			c.Set("access_type", "console")
			c.Next()
			return
		}

		// Check for store subdomain ({slug}.{baseDomain})
		if strings.HasSuffix(host, "."+baseDomain) {
			slug := strings.TrimSuffix(host, "."+baseDomain)
			// Exclude special subdomains
			if slug != "www" && slug != "app" && slug != "console" && slug != "api" && slug != "stores" {
				c.Set("access_type", "subdomain")
				c.Set("store_slug", slug)
				c.Next()
				return
			}
		}

		// Check for custom domain
		// Check for custom domain
		var store models.Store
		if database.GetDB().Where("custom_domain = ? AND domain_verified = true", host).First(&store).Error == nil {
			// Check if User Plan allows Custom Domain
			var user models.User
			allowed := false
			if err := database.GetDB().Preload("Plan").First(&user, store.UserID).Error; err == nil {
				if user.Plan != nil && user.Plan.AllowCustomDomain {
					allowed = true
				}
			}

			if allowed {
				c.Set("access_type", "custom_domain")
				c.Set("store_slug", store.Slug)
				c.Set("store_id", store.ID)
				c.Next()
				return
			}
		}

		// Unknown domain - treat as main
		c.Set("access_type", "unknown")
		c.Next()
	}
}

// RequireSubdomain checks if request is from a specific subdomain
func RequireSubdomain(subdomain string, baseDomain string) gin.HandlerFunc {
	return func(c *gin.Context) {
		host := c.Request.Host
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		expected := subdomain + "." + baseDomain
		if host != expected {
			c.Redirect(302, "https://"+expected+c.Request.URL.Path)
			c.Abort()
			return
		}
		c.Next()
	}
}

// StoreFromDomain - middleware to load store from subdomain/custom domain
func StoreFromDomain() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessType, exists := c.Get("access_type")
		if !exists {
			c.Next()
			return
		}

		if accessType == "subdomain" || accessType == "custom_domain" {
			slug, _ := c.Get("store_slug")
			if slugStr, ok := slug.(string); ok && slugStr != "" {
				var store models.Store
				if err := database.GetDB().Where("slug = ?", slugStr).First(&store).Error; err == nil {
					c.Set("store", store)
				}
			}
		}
		c.Next()
	}
}
