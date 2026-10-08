package services

import (
	"crypto/rand"
	"time"

	"sello/internal/database"
	"sello/internal/models"

	"github.com/google/uuid"
)

// GenerateOTP creates a 6-digit code and saves it to DB
func GenerateOTP(userID uuid.UUID, otpType string) (string, error) {
	// Generate 6 digit random number
	b := make([]byte, 6)
	rand.Read(b)
	for i := 0; i < 6; i++ {
		b[i] = table[int(b[i])%len(table)]
	}
	code := string(b)

	otp := models.OTP{
		UserID:    userID,
		Code:      code,
		Type:      otpType,
		ExpiresAt: time.Now().Add(15 * time.Minute), // Valid for 15 mins
		CreatedAt: time.Now(),
		Used:      false,
	}

	if err := database.GetDB().Create(&otp).Error; err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTP checks if code is valid and unused
func VerifyOTP(userID uuid.UUID, code, otpType string) bool {
	var otp models.OTP
	err := database.GetDB().Where("user_id = ? AND code = ? AND type = ? AND used = ? AND expires_at > ?", 
		userID, code, otpType, false, time.Now()).First(&otp).Error

	if err != nil {
		return false
	}

	// Mark as used
	otp.Used = true
	database.GetDB().Save(&otp)

	return true
}

var table = [...]byte{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'}
