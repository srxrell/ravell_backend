package handlers

import (
	"net/http"
	"strconv"
	"ravell_backend/models"
	"ravell_backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetMyProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	// ✅ ИСПРАВЛЕНИЕ: Получаем user_id как строку и конвертируем
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Находим пользователя с профилем
	var user models.User
	if err := db.Preload("Profile").First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"avatar":     user.Profile.Avatar,
		"bio":        user.Profile.Bio,
		"is_verified": user.Profile.IsVerified,
		"created_at": user.CreatedAt,
	})
}

func UpdateProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	// ✅ ИСПРАВЛЕНИЕ: Получаем user_id как строку и конвертируем
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var req struct {
		Username  string `json:"username"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Avatar    string `json:"avatar"`
		Bio       string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем уникальность username и email
	var existingUser models.User
	if err := db.Where("(username = ? OR email = ?) AND id != ?", req.Username, req.Email, uint(userID)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	// Обновляем пользователя
	var user models.User
	if err := db.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Обновляем поля пользователя
	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	user.FirstName = req.FirstName
	user.LastName = req.LastName

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Обновляем профиль
	var profile models.Profile
	if err := db.Where("user_id = ?", uint(userID)).First(&profile).Error; err != nil {
		// Если профиля нет, создаем новый
		profile = models.Profile{
			UserID:     uint(userID),
			IsVerified: true,
		}
	}

	if req.Avatar != "" {
		profile.Avatar = req.Avatar
	}
	if req.Bio != "" {
		profile.Bio = req.Bio
	}

	if err := db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar":     profile.Avatar,
			"bio":        profile.Bio,
			"is_verified": profile.IsVerified,
		},
	})
}

// UpdateProfileWithImage обрабатывает обновление профиля с загрузкой аватара
func UpdateProfileWithImage(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	// ✅ ИСПРАВЛЕНИЕ: Получаем user_id как строку и конвертируем
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr.(string), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	// Парсим multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil { // 10MB max
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}

	// Получаем текстовые поля
	username := c.Request.FormValue("username")
	email := c.Request.FormValue("email")
	firstName := c.Request.FormValue("first_name")
	lastName := c.Request.FormValue("last_name")
	bio := c.Request.FormValue("bio")

	// Проверяем уникальность username и email
	var existingUser models.User
	if err := db.Where("(username = ? OR email = ?) AND id != ?", username, email, uint(userID)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username or email already exists"})
		return
	}

	// Обновляем пользователя
	var user models.User
	if err := db.First(&user, uint(userID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Обновляем поля пользователя
	if username != "" {
		user.Username = username
	}
	if email != "" {
		user.Email = email
	}
	user.FirstName = firstName
	user.LastName = lastName

	if err := db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	// Обрабатываем загрузку аватара
	avatarURL := ""
	file, header, err := c.Request.FormFile("avatar")
	if err == nil {
		defer file.Close()
		
		// Здесь должна быть логика загрузки файла в облачное хранилище
		// Для примера возвращаем временный URL
		avatarURL = "/uploads/" + header.Filename
		
		// В реальном приложении используйте:
		// - AWS S3
		// - Google Cloud Storage  
		// - Cloudinary
		// - Или сохраняйте локально
	}

	// Обновляем профиль
	var profile models.Profile
	if err := db.Where("user_id = ?", uint(userID)).First(&profile).Error; err != nil {
		// Если профиля нет, создаем новый
		profile = models.Profile{
			UserID:     uint(userID),
			IsVerified: true,
		}
	}

	if avatarURL != "" {
		profile.Avatar = avatarURL
	}
	if bio != "" {
		profile.Bio = bio
	}

	if err := db.Save(&profile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user": gin.H{
			"id":         user.ID,
			"username":   user.Username,
			"email":      user.Email,
			"first_name": user.FirstName,
			"last_name":  user.LastName,
			"avatar":     profile.Avatar,
			"bio":        profile.Bio,
			"is_verified": profile.IsVerified,
		},
	})
}