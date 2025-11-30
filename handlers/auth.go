package handlers

import (
	"fmt"
	"net/http"
	"os"
	"ravell_backend/models"
	"ravell_backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Register(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка существующего пользователя
	var existingUser models.User
	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	// Хеширование пароля
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создание пользователя
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Создание профиля
	profile := models.Profile{UserID: user.ID}
	if err := db.Create(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
		return
	}

	// Генерация и отправка OTP
	otp, err := utils.SaveOTP(db, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	fmt.Printf("✅ OTP generated for user %s (%s): %s\n", user.Username, user.Email, otp)

	// Отправка email
	if err := utils.SendOTPEmail(user.Email, user.Username, otp); err != nil {
		fmt.Printf("❌ Email sending failed: %v\n", err)
		
		// Временно возвращаем OTP для разработки
		c.JSON(http.StatusCreated, gin.H{
			"message":    "User registered successfully. OTP generated but email failed",
			"user_id":    user.ID,
			"otp":        otp, // Только для разработки!
			"debug_info": "Check email configuration",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully. OTP sent to email",
		"user_id": user.ID,
	})
}

func Login(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Preload("Profile").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.Profile.IsVerified {
		c.JSON(http.StatusForbidden, gin.H{"error": "Account not verified"})
		return
	}

	tokens, err := utils.GenerateJWTToken(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Login successful",
		"user_id":  user.ID,
		"username": user.Username,
		"tokens":   tokens,
	})
}

func VerifyOTP(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var req struct {
		UserID uint   `json:"user_id" binding:"required"`
		OTP    string `json:"otp" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	success, err := utils.VerifyOTP(db, req.UserID, req.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !success {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP verification failed"})
		return
	}

	tokens, err := utils.GenerateJWTToken(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Account verified successfully",
		"tokens":  tokens,
	})
}

func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tokens, err := utils.RefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Token refreshed successfully",
		"tokens":  tokens,
	})
}

func ResendOTP(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	var req struct {
		UserID uint `json:"user_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := db.Preload("Profile").First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Генерация нового OTP
	otp, err := utils.SaveOTP(db, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate OTP"})
		return
	}

	fmt.Printf("✅ New OTP generated for user %s (%s): %s\n", user.Username, user.Email, otp)

	// Отправка email
	if err := utils.SendOTPEmail(user.Email, user.Username, otp); err != nil {
		fmt.Printf("❌ Email sending failed: %v\n", err)
		
		// Временно возвращаем OTP для разработки
		c.JSON(http.StatusOK, gin.H{
			"message":    "OTP regenerated but email failed",
			"otp":        otp, // Только для разработки!
			"debug_info": "Check email configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP resent successfully",
	})
}

// TestEmailConfig - для проверки настроек email
func TestEmailConfig(c *gin.Context) {
	config := map[string]string{
		"SMTP_HOST":      os.Getenv("SMTP_HOST"),
		"SMTP_PORT":      os.Getenv("SMTP_PORT"),
		"SMTP_USER":      os.Getenv("SMTP_USER"),
		"FROM_EMAIL":     os.Getenv("FROM_EMAIL"),
		"SMTP_PASS_set":  fmt.Sprintf("%v", os.Getenv("SMTP_PASS") != ""),
	}

	c.JSON(http.StatusOK, gin.H{
		"email_config": config,
		"message":      "Check if SMTP_PASS is set correctly",
	})
}