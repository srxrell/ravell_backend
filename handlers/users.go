package handlers

import (
	"net/http"
	"strconv"
	"ravell_backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserProfile(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	result := db.Preload("Profile").First(&user, userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Статистика пользователя
	var storiesCount int64
	var followersCount int64
	var followingCount int64

	db.Model(&models.Story{}).Where("user_id = ?", userID).Count(&storiesCount)
	db.Model(&models.Subscription{}).Where("following_id = ?", userID).Count(&followersCount)
	db.Model(&models.Subscription{}).Where("follower_id = ?", userID).Count(&followingCount)

	response := gin.H{
		"user": user,
		"stats": gin.H{
			"stories_count":   storiesCount,
			"followers_count": followersCount,
			"following_count": followingCount,
		},
	}

	c.JSON(http.StatusOK, response)
}


func GetFollowers(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var followers []models.Subscription
	result := db.Where("following_id = ?", userID).Find(&followers)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followers"})
		return
	}

	// Получаем данные пользователей
	var followersData []gin.H
	for _, sub := range followers {
		var user models.User
		if err := db.Preload("Profile").First(&user, sub.FollowerID).Error; err != nil {
			continue
		}
		followersData = append(followersData, gin.H{
			"user": user,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"followers": followersData,
		"count":     len(followersData),
	})
}

func GetFollowing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	
	userID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var following []models.Subscription
	result := db.Where("follower_id = ?", userID).Find(&following)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following"})
		return
	}

	// Получаем данные пользователей
	var followingData []gin.H
	for _, sub := range following {
		var user models.User
		if err := db.Preload("Profile").First(&user, sub.FollowingID).Error; err != nil {
			continue
		}
		followingData = append(followingData, gin.H{
			"user": user,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"following": followingData,
		"count":     len(followingData),
	})
}

func FollowUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	currentUserID := c.MustGet("user_id").(uint)
	
	userIDToFollow, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Нельзя подписаться на себя
	if uint(userIDToFollow) == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
		return
	}

	// Проверяем существование пользователя
	var userToFollow models.User
	if err := db.First(&userToFollow, userIDToFollow).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Проверяем существование подписки
	var existingSubscription models.Subscription
	if err := db.Where("follower_id = ? AND following_id = ?", currentUserID, userIDToFollow).
		First(&existingSubscription).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already following this user"})
		return
	}

	// Создаем подписку
	subscription := models.Subscription{
		FollowerID:  currentUserID,
		FollowingID: uint(userIDToFollow),
	}

	if err := db.Create(&subscription).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Successfully followed user",
		"user":    userToFollow,
	})
}

func UnfollowUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	currentUserID := c.MustGet("user_id").(uint)
	
	userIDToUnfollow, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Удаляем подписку
	result := db.Where("follower_id = ? AND following_id = ?", currentUserID, userIDToUnfollow).
		Delete(&models.Subscription{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unfollow user"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not following this user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Successfully unfollowed user",
	})
}
