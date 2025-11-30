package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"ravell_backend/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMyProfile получает профиль текущего пользователя
func GetMyProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := db.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Получаем статистику
	var stats struct {
		StoriesCount   int64 `json:"stories_count"`
		FollowersCount int64 `json:"followers_count"`
		FollowingCount int64 `json:"following_count"`
	}

	db.Model(&models.Story{}).Where("user_id = ?", user.ID).Count(&stats.StoriesCount)
	db.Model(&models.Subscription{}).Where("following_id = ?", user.ID).Count(&stats.FollowersCount)
	db.Model(&models.Subscription{}).Where("follower_id = ?", user.ID).Count(&stats.FollowingCount)

	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"profile": user.Profile,
		"stats":   stats,
	})
}

// UpdateProfile обновляет профиль (без изображения)
func UpdateProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req struct {
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Bio       string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем пользователя
	if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Обновляем профиль
	if err := db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"bio": req.Bio,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Получаем обновленные данные
	var user models.User
	if err := db.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
		"profile": user.Profile,
	})
}

// UpdateProfileWithImage обновляет профиль с загрузкой аватара
func UpdateProfileWithImage(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Парсим multipart форму
	if err := c.Request.ParseMultipartForm(10 << 20); // 10MB
	err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	// Получаем текстовые поля
	firstName := c.Request.FormValue("first_name")
	lastName := c.Request.FormValue("last_name")
	bio := c.Request.FormValue("bio")

	// Обновляем пользователя
	updates := map[string]interface{}{}
	if firstName != "" {
		updates["first_name"] = firstName
	}
	if lastName != "" {
		updates["last_name"] = lastName
	}

	if len(updates) > 0 {
		if err := db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// Обрабатываем загрузку аватара
	avatarURL := ""
	file, header, err := c.Request.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// Создаем папку media если её нет
		if err := os.MkdirAll("media/avatars", 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create media directory"})
			return
		}

		// Генерируем уникальное имя файла
		ext := filepath.Ext(header.Filename)
		filename := fmt.Sprintf("avatar_%d_%d%s", userID, time.Now().Unix(), ext)
		filePath := filepath.Join("media", "avatars", filename)

		// Сохраняем файл
		if err := c.SaveUploadedFile(header, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
			return
		}

		// Формируем URL для доступа к файлу
		avatarURL = fmt.Sprintf("/media/avatars/%s", filename)
	}

	// Обновляем профиль
	profileUpdates := map[string]interface{}{
		"bio": bio,
	}
	if avatarURL != "" {
		profileUpdates["avatar"] = avatarURL
	}

	if err := db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(profileUpdates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Получаем обновленные данные
	var user models.User
	if err := db.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
		"profile": user.Profile,
	})
}

// GetUserProfile получает профиль любого пользователя по ID
func GetUserProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID := c.Param("id")
	
	var user models.User
	if err := db.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Получаем статистику
	var stats struct {
		StoriesCount   int64 `json:"stories_count"`
		FollowersCount int64 `json:"followers_count"`
		FollowingCount int64 `json:"following_count"`
	}

	db.Model(&models.Story{}).Where("user_id = ?", user.ID).Count(&stats.StoriesCount)
	db.Model(&models.Subscription{}).Where("following_id = ?", user.ID).Count(&stats.FollowersCount)
	db.Model(&models.Subscription{}).Where("follower_id = ?", user.ID).Count(&stats.FollowingCount)

	// Получаем последние 10 историй пользователя
	var stories []models.Story
	db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(10).Find(&stories)

	// Проверяем, подписан ли текущий пользователь на этого пользователя
	isFollowing := false
	if currentUserID, exists := c.Get("user_id"); exists {
		var subscription models.Subscription
		if err := db.Where("follower_id = ? AND following_id = ?", currentUserID, user.ID).First(&subscription).Error; err == nil {
			isFollowing = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"profile":      user.Profile,
		"stats":        stats,
		"stories":      stories,
		"is_following": isFollowing,
	})
}