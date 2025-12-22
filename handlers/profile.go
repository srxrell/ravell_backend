package handlers

import (
	"fmt"
	"go_stories_api/models"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

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
		db.Model(&models.User{}).
			Where("id = ?", userID).
			Updates(userUpdates)
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

		// гарантируем папку
		avatarsDir := "./media/avatars"
		if err := os.MkdirAll(avatarsDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create media dir"})
			return
		}

		filename := fmt.Sprintf(
			"avatar_%d_%d%s",
			userID,
			time.Now().Unix(),
			filepath.Ext(header.Filename),
		)

		fullPath := filepath.Join(avatarsDir, filename)

		out, err := os.Create(fullPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, file); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write avatar"})
			return
		}

		// URL, который будет жрать фронт
		profileUpdates["avatar"] = "/media/avatars/" + filename
	}

	if len(profileUpdates) > 0 {
		db.Model(&models.Profile{}).
			Where("user_id = ?", userID).
			Updates(profileUpdates)
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
