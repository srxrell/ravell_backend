package handlers

import (
	"net/http"
	"ravell_backend/models"
	"ravell_backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// func Register(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
	
// 	var req struct {
// 		Username string `json:"username" binding:"required"`
// 		Email    string `json:"email" binding:"required,email"`
// 		Password string `json:"password" binding:"required,min=6"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Проверка существующего пользователя
// 	var existingUser models.User
// 	if err := db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
// 		return
// 	}

// 	// Хеширование пароля
// 	hashedPassword, err := utils.HashPassword(req.Password)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
// 		return
// 	}

// 	// Создание пользователя
// 	user := models.User{
// 		Username: req.Username,
// 		Email:    req.Email,
// 		Password: hashedPassword,
// 	}

// 	if err := db.Create(&user).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
// 		return
// 	}

// 	// Создание профиля СРАЗУ ВЕРИФИЦИРОВАННЫМ
// 	profile := models.Profile{
// 		UserID:      user.ID,
// 		IsVerified:  true, // Сразу верифицируем
// 		OtpCode:     "",   // Пустой OTP
// 		OtpCreatedAt: time.Time{},
// 	}
	
// 	if err := db.Create(&profile).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create profile"})
// 		return
// 	}

// 	// ГЕНЕРАЦИЯ ТОКЕНОВ СРАЗУ (без OTP)
// 	tokens, err := utils.GenerateJWTToken(user.ID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"message":  "User registered successfully",
// 		"user_id":  user.ID,
// 		"username": user.Username,
// 		"tokens":   tokens, // Сразу возвращаем токены
// 	})
// }

// func Login(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
	
// 	var req struct {
// 		Username string `json:"username" binding:"required"`
// 		Password string `json:"password" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	var user models.User
// 	if err := db.Preload("Profile").Where("username = ?", req.Username).First(&user).Error; err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	if !utils.CheckPasswordHash(req.Password, user.Password) {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
// 		return
// 	}

// 	tokens, err := utils.GenerateJWTToken(user.ID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message":  "Login successful",
// 		"user_id":  user.ID,
// 		"username": user.Username,
// 		"tokens":   tokens,
// 	})
// }

// func RefreshToken(c *gin.Context) {
// 	var req struct {
// 		RefreshToken string `json:"refresh_token" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	tokens, err := utils.RefreshToken(req.RefreshToken)
// 	if err != nil {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Token refreshed successfully",
// 		"tokens":  tokens,
// 	})
// }

// func DeleteAccount(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
	
// 	userID, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
// 		return
// 	}

// 	var req struct {
// 		Password string `json:"password" binding:"required"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	// Находим пользователя
// 	var user models.User
// 	if err := db.First(&user, userID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	// Проверяем пароль
// 	if !utils.CheckPasswordHash(req.Password, user.Password) {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
// 		return
// 	}

// 	// Удаляем в транзакции
// 	err := db.Transaction(func(tx *gorm.DB) error {
// 		// Удаляем профиль
// 		if err := tx.Where("user_id = ?", userID).Delete(&models.Profile{}).Error; err != nil {
// 			return err
// 		}
		
// 		// Удаляем связанные данные
// 		tx.Where("user_id = ?", userID).Delete(&models.Story{})
// 		tx.Where("user_id = ?", userID).Delete(&models.Comment{})
// 		tx.Where("user_id = ?", userID).Delete(&models.Like{})
// 		tx.Where("user_id = ?", userID).Delete(&models.Subscription{})
// 		tx.Where("user_id = ?", userID).Delete(&models.NotInterested{})
		
// 		// Удаляем самого пользователя
// 		if err := tx.Delete(&user).Error; err != nil {
// 			return err
// 		}
		
// 		return nil
// 	})

// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete account"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Account deleted successfully",
// 	})
// }

// // GetUserStories получает истории пользователя
// func GetUserStories(c *gin.Context) {
// 	db := c.MustGet("db").(*gorm.DB)
	
// 	userID := c.Param("id")
	
// 	var user models.User
// 	if err := db.First(&user, userID).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
// 		return
// 	}

// 	var stories []models.Story
// 	if err := db.Preload("User").Preload("Hashtags.Hashtag").Where("user_id = ?", userID).Order("created_at DESC").Find(&stories).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stories"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"stories": stories,
// 	})
// }

// GetFollowers получает подписчиков пользователя
func GetFollowers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID := c.Param("id")
	
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var followers []models.Subscription
	if err := db.Preload("Follower.Profile").Where("following_id = ?", userID).Find(&followers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followers"})
		return
	}

	// Форматируем ответ
	var result []gin.H
	for _, sub := range followers {
		result = append(result, gin.H{
			"id":        sub.Follower.ID,
			"username":  sub.Follower.Username,
			"avatar":    sub.Follower.Profile.Avatar,
			"bio":       sub.Follower.Profile.Bio,
			"followed_at": sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": result,
	})
}

// GetFollowing получает подписки пользователя
func GetFollowing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID := c.Param("id")
	
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var following []models.Subscription
	if err := db.Preload("Following.Profile").Where("follower_id = ?", userID).Find(&following).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following"})
		return
	}

	// Форматируем ответ
	var result []gin.H
	for _, sub := range following {
		result = append(result, gin.H{
			"id":        sub.Following.ID,
			"username":  sub.Following.Username,
			"avatar":    sub.Following.Profile.Avatar,
			"bio":       sub.Following.Profile.Bio,
			"followed_at": sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"following": result,
	})
}

// FollowUser подписывается на пользователя
func FollowUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("id")
	
	// Нельзя подписаться на себя
	if currentUserID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
		return
	}

	// Проверяем существование целевого пользователя
	var targetUser models.User
	if err := db.First(&targetUser, targetUserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Проверяем, не подписаны ли уже
	var existingSubscription models.Subscription
	if err := db.Where("follower_id = ? AND following_id = ?", currentUserID, targetUserID).First(&existingSubscription).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Already following this user"})
		return
	}

	// Создаем подписку
	subscription := models.Subscription{
		FollowerID:  currentUserID.(uint),
		FollowingID: targetUser.ID,
	}

	if err := db.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully followed user",
	})
}

// UnfollowUser отписывается от пользователя
func UnfollowUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	targetUserID := c.Param("id")

	// Удаляем подписку
	result := db.Where("follower_id = ? AND following_id = ?", currentUserID, targetUserID).Delete(&models.Subscription{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not following this user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unfollowed user",
	})
}