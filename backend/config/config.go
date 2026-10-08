package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	SupabaseURL       string
	SupabaseKey       string
	SupabaseJWTSecret string
	JWTSecret         string // For admin tokens
	DatabaseURL       string
	Port              string
	GinMode           string
	BaseDomain        string // sello.id
	// Cloudinary
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
}

var AppConfig *Config

func Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		SupabaseURL:         getEnv("SUPABASE_URL", ""),
		SupabaseKey:         getEnv("SUPABASE_KEY", ""),
		SupabaseJWTSecret:   getEnv("SUPABASE_JWT_SECRET", ""),
		JWTSecret:           getEnv("JWT_SECRET", getEnv("SUPABASE_JWT_SECRET", "defaultsecret")),
		DatabaseURL:         getEnv("DATABASE_URL", ""),
		Port:                getEnv("PORT", "8080"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		BaseDomain:          getEnv("BASE_DOMAIN", "sello.id"),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME", ""),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY", ""),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET", ""),
		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
