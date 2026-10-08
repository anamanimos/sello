package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"sello/config"
	"sello/internal/database"
	"sello/internal/models"
	"sello/internal/services"
)

var googleOauthConfig *oauth2.Config

func InitGoogleAuth() {
	// Ensure config is loaded
	if config.AppConfig == nil {
		config.Load()
	}

	googleOauthConfig = &oauth2.Config{
		RedirectURL:  config.AppConfig.GoogleRedirectURL,
		ClientID:     config.AppConfig.GoogleClientID,
		ClientSecret: config.AppConfig.GoogleClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

// GoogleLogin - GET /api/auth/google/login
func GoogleLogin(c *gin.Context) {
	if googleOauthConfig == nil {
		InitGoogleAuth()
	}
	// Generate random state (should verify in callback, but keep simple for now)
	url := googleOauthConfig.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback - GET /api/auth/google/callback
func GoogleCallback(c *gin.Context) {
	if googleOauthConfig == nil {
		InitGoogleAuth()
	}

	code := c.Query("code")
	if code == "" {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=access_denied")
		return
	}

	token, err := googleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=token_exchange_failed")
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=user_info_failed")
		return
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=read_body_failed")
		return
	}

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
	}

	if err := json.Unmarshal(content, &googleUser); err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=parse_json_failed")
		return
	}

	// Upsert User logic
	db := database.GetDB()
	var user models.User
	result := db.Where("email = ?", googleUser.Email).First(&user)

	if result.Error != nil {
		// New User
		// Fetch Default Plan ID
		var defaultPlanID uint = 1 // Fallback to 1 (Free)
		var planSetting models.SystemSetting
		if err := db.Where("key = ?", "default_register_plan_id").First(&planSetting).Error; err == nil {
			// Parse
			if val, err := strconv.Atoi(planSetting.Value); err == nil {
				defaultPlanID = uint(val)
			}
		}

		user = models.User{
			ID:              uuid.New(),
			Name:            googleUser.Name,
			Email:           googleUser.Email,
			IsEmailVerified: true,
			IsPhoneVerified: true,
			PlanID:          &defaultPlanID,
		}
		if err := db.Create(&user).Error; err != nil {
			c.Redirect(http.StatusTemporaryRedirect, "/login?error=create_user_failed")
			return
		}
	} else {
		// Update existing user
		user.IsEmailVerified = true
         // If name is empty, update it
         if user.Name == "" {
             user.Name = googleUser.Name
             db.Save(&user)
         } else {
             db.Save(&user) // Save verification status
         }
	}

	// Generate JWT
	tokenString, err := services.GenerateToken(user.ID.String())
	if err != nil {
		c.Redirect(http.StatusTemporaryRedirect, "/login?error=token_generation_failed")
		return
	}

	// Redirect to callback handler page (frontend stores token)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("/auth/callback?token=%s", tokenString))
}
