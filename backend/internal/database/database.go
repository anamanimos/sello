package database

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"sello/config"
)

var DB *gorm.DB

func Connect() {
	var err error

	// Use individual connection params if available (recommended for special chars in password)
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")

	var dsn string
	if host != "" && user != "" && password != "" {
		// Determine SSL mode based on host
		sslmode := "require"
		if host == "localhost" || host == "127.0.0.1" {
			sslmode = "disable"
		}

		// Use DSN format that doesn't require URL encoding
		dsn = "host=" + host + " user=" + user + " password=" + password + " dbname=" + dbname + " port=" + port + " sslmode=" + sslmode
	} else {
		// Fallback to DATABASE_URL
		dsn = config.AppConfig.DatabaseURL
		if dsn == "" {
			log.Fatal("DATABASE_URL or DB_* environment variables are not set")
		}
	}

	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Disable implicit prepared statements
	}), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: false, // Explicitly disable prepared statements cache
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connected successfully")
}

func GetDB() *gorm.DB {
	return DB
}
