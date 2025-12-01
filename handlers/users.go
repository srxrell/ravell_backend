package handlers

import (
	"net/http"
	"go_stories_api/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserStories получает истории пользователя
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

	var subscriptions []models.Subscription
	if err := db.Where("following_id = ?", userID).Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch followers"})
		return
	}

	// Получаем информацию о подписчиках
	var result []gin.H
	for _, sub := range subscriptions {
		var follower models.User
		if err := db.Preload("Profile").First(&follower, sub.FollowerID).Error; err != nil {
			continue
		}
		
		result = append(result, gin.H{
			"id":         follower.ID,
			"username":   follower.Username,
			"avatar":     follower.Profile.Avatar,
			"bio":        follower.Profile.Bio,
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

	var subscriptions []models.Subscription
	if err := db.Where("follower_id = ?", userID).Find(&subscriptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch following"})
		return
	}

	// Получаем информацию о подписках
	var result []gin.H
	for _, sub := range subscriptions {
		var followingUser models.User
		if err := db.Preload("Profile").First(&followingUser, sub.FollowingID).Error; err != nil {
			continue
		}
		
		result = append(result, gin.H{
			"id":         followingUser.ID,
			"username":   followingUser.Username,
			"avatar":     followingUser.Profile.Avatar,
			"bio":        followingUser.Profile.Bio,
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