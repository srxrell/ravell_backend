package handlers

import (
	"fmt"
	"go_stories_api/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"bytes"
	"mime/multipart"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetMyProfile
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

	var achievement models.Achievement
	if err := db.Where("key = ?", "early_access").First(&achievement).Error; err == nil {
		var ua models.UserAchievement
		if err := db.Where("user_id = ? AND achievement_id = ?", user.ID, achievement.ID).First(&ua).Error; err != nil {
			ua = models.UserAchievement{
				UserID:        user.ID,
				AchievementID: achievement.ID,
				Progress:      1.0,
				Unlocked:      true,
			}
			db.Create(&ua)
		}
	}

	var stats struct {
		StoriesCount   int64 `json:"stories_count"`
		FollowersCount int64 `json:"followers_count"`
		FollowingCount int64 `json:"following_count"`
	}

	db.Model(&models.Story{}).Where("user_id = ?", user.ID).Count(&stats.StoriesCount)
	db.Model(&models.Subscription{}).Where("following_id = ?", user.ID).Count(&stats.FollowersCount)
	db.Model(&models.Subscription{}).Where("follower_id = ?", user.ID).Count(&stats.FollowingCount)

	earlyCutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	isEarly := user.Profile.IsEarly || user.CreatedAt.Before(earlyCutoff)

	c.JSON(http.StatusOK, gin.H{
		"user":     user,
		"profile":  user.Profile,
		"stats":    stats,
		"is_early": isEarly,
	})
}

// UpdateProfile
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

	db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"first_name": req.FirstName,
		"last_name":  req.LastName,
	})

	db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(map[string]interface{}{
		"bio": req.Bio,
	})

	var user models.User
	db.Preload("Profile").First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Profile updated successfully",
		"user":     user,
		"profile":  user.Profile,
		"is_early": user.Profile.IsEarly,
	})
}

// UpdateProfileWithImage
// UpdateProfileWithImage загружает аватарку сразу на 0x0.st
func UpdateProfileWithImage(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form"})
		return
	}

	firstName := c.Request.FormValue("first_name")
	lastName := c.Request.FormValue("last_name")
	bio := c.Request.FormValue("bio")

	// --- обновляем user ---
	userUpdates := map[string]interface{}{}
	if firstName != "" {
		userUpdates["first_name"] = firstName
	}
	if lastName != "" {
		userUpdates["last_name"] = lastName
	}
	if len(userUpdates) > 0 {
		db.Model(&models.User{}).Where("id = ?", userID).Updates(userUpdates)
	}

	// --- profile ---
	profileUpdates := map[string]interface{}{}
	if bio != "" {
		profileUpdates["bio"] = bio
	}

	// --- файл ---
	file, header, err := c.Request.FormFile("avatar")
	if err == nil {
		defer file.Close()

		// создаём временный файл
		tmpFilePath := fmt.Sprintf("/tmp/avatar_%d%s", time.Now().UnixNano(), filepath.Ext(header.Filename))
		out, err := os.Create(tmpFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create temp file"})
			return
		}
		if _, err := io.Copy(out, file); err != nil {
			out.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write temp file"})
			return
		}
		out.Close()

		// --- загружаем на 0x0.st ---
		avatarURL, err := uploadTo0x0(tmpFilePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload avatar"})
			return
		}

		// удаляем временный файл
		os.Remove(tmpFilePath)

		avatarURL = strings.TrimSpace(avatarURL)

		profileUpdates["avatar"] = avatarURL
	}

	if len(profileUpdates) > 0 {
		db.Model(&models.Profile{}).Where("user_id = ?", userID).Updates(profileUpdates)
	}

	var user models.User
	db.Preload("Profile").First(&user, userID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Profile updated successfully",
		"user":     user,
		"profile":  user.Profile,
		"is_early": user.Profile.IsEarly,
	})
}

// uploadTo0x0 загружает файл на 0x0.st через чистый HTTP POST
func uploadTo0x0(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(file.Name()))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	writer.Close()
	req, err := http.NewRequest("POST", "https://0x0.st", body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// ✅ FIX: Add User-Agent to avoid "User agent not allowed" error
	req.Header.Set("User-Agent", "Ravell/1.0 (Go-http-client)") 
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err	
	}
	// ✅ FIX: Check for success status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("upload failed: %s", string(respData))
	}
	
	cleanURL := strings.TrimSpace(string(respData))
	return cleanURL, nil
}



// GetUserProfile
func GetUserProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.Param("id")

	var user models.User
	if err := db.Preload("Profile").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var stats struct {
		StoriesCount   int64 `json:"stories_count"`
		FollowersCount int64 `json:"followers_count"`
		FollowingCount int64 `json:"following_count"`
	}

	db.Model(&models.Story{}).Where("user_id = ?", user.ID).Count(&stats.StoriesCount)
	db.Model(&models.Subscription{}).Where("following_id = ?", user.ID).Count(&stats.FollowersCount)
	db.Model(&models.Subscription{}).Where("follower_id = ?", user.ID).Count(&stats.FollowingCount)

	earlyCutoff := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	isEarly := user.Profile.IsEarly || user.CreatedAt.Before(earlyCutoff)

	var stories []models.Story
	db.Where("user_id = ?", user.ID).Order("created_at DESC").Limit(10).Find(&stories)

	isFollowing := false
	if currentUserID, exists := c.Get("user_id"); exists {
		var sub models.Subscription
		if err := db.Where("follower_id = ? AND following_id = ?", currentUserID, user.ID).First(&sub).Error; err == nil {
			isFollowing = true
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user":         user,
		"profile":      user.Profile,
		"stats":        stats,
		"stories":      stories,
		"is_following": isFollowing,
		"is_early":     isEarly,
	})
}
