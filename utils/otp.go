package utils

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"ravell_backend/models"
	"time"

	"gorm.io/gorm"
)

func GenerateOTP() string {
	const digits = "0123456789"
	const length = 6
	otp := make([]byte, length)
	
	for i := range otp {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		otp[i] = digits[num.Int64()]
	}
	
	return string(otp)
}

func SaveOTP(db *gorm.DB, userID uint) (string, error) {
	otp := GenerateOTP()
	
	var profile models.Profile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return "", fmt.Errorf("profile not found for user %d: %v", userID, err)
	}
	
	profile.OtpCode = otp
	profile.OtpCreatedAt = time.Now()
	profile.IsVerified = false
	
	if err := db.Save(&profile).Error; err != nil {
		return "", fmt.Errorf("failed to save OTP: %v", err)
	}
	
	fmt.Printf("💾 OTP saved for user %d: %s\n", userID, otp)
	return otp, nil
}

func VerifyOTP(db *gorm.DB, userID uint, enteredOTP string) (bool, error) {
	var profile models.Profile
	if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
		return false, fmt.Errorf("profile not found: %v", err)
	}
	
	if time.Since(profile.OtpCreatedAt) > 15*time.Minute {
		return false, fmt.Errorf("OTP expired")
	}
	
	if profile.OtpCode != enteredOTP {
		return false, fmt.Errorf("invalid OTP")
	}
	
	profile.IsVerified = true
	profile.OtpCode = ""
	profile.OtpCreatedAt = time.Time{}
	
	if err := db.Save(&profile).Error; err != nil {
		return false, fmt.Errorf("failed to update profile: %v", err)
	}
	
	return true, nil
}