package handlers

import (
	"go_stories_api/models"
	"go_stories_api/wsservice"
	"net/http"
	"strconv"
	"time"

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

	var subscriptions []models.Subscription
	db.Where("following_id = ?", userID).Find(&subscriptions)

	var result []gin.H
	for _, sub := range subscriptions {
		var follower models.User
		db.Preload("Profile").First(&follower, sub.FollowerID)

		// Проверяем, подписан ли текущий пользователь на этого юзера
		currentUserID, _ := c.Get("user_id")
		var isFollowing bool
		if currentUserID != nil {
			var exists int64
			db.Model(&models.Subscription{}).
				Where("follower_id = ? AND following_id = ?", currentUserID, follower.ID).
				Count(&exists)
			isFollowing = exists > 0
		}

		result = append(result, gin.H{
			"user": gin.H{
				"id":       follower.ID,
				"username": follower.Username,
				"bio":      follower.Profile.Bio,
				"avatar":   follower.Profile.Avatar,
			},
			"is_following": isFollowing,
			"followed_at":  sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"followers": result})
}

func GetFollowing(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userID := c.Param("id")

	var subscriptions []models.Subscription
	db.Where("follower_id = ?", userID).Find(&subscriptions)

	var result []gin.H
	for _, sub := range subscriptions {
		var followingUser models.User
		db.Preload("Profile").First(&followingUser, sub.FollowingID)

		currentUserID, _ := c.Get("user_id")
		var isFollowing bool
		if currentUserID != nil {
			var exists int64
			db.Model(&models.Subscription{}).
				Where("follower_id = ? AND following_id = ?", currentUserID, followingUser.ID).
				Count(&exists)
			isFollowing = exists > 0
		}

		result = append(result, gin.H{
			"user": gin.H{
				"id":       followingUser.ID,
				"username": followingUser.Username,
				"bio":      followingUser.Profile.Bio,
				"avatar":      followingUser.Profile.Avatar,
			},
			"is_following": isFollowing,
			"followed_at":  sub.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"following": result})
}

func FollowUser(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    followerID := c.MustGet("user_id").(uint)

    followeeIDStr := c.Param("id")
    followeeID64, err := strconv.ParseUint(followeeIDStr, 10, 64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
        return
    }
    followeeID := uint(followeeID64)

    if followerID == followeeID {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot follow yourself"})
        return
    }

    var followee models.User
    if err := db.First(&followee, followeeID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User to follow not found"})
        return
    }

    var existingSub models.Subscription
    if err := db.Where("follower_id = ? AND following_id = ?", followerID, followeeID).First(&existingSub).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Already following this user"})
        return
    }

    subscription := models.Subscription{
        FollowerID:  followerID,
        FollowingID: followeeID,
    }
    if err := db.Create(&subscription).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow"})
        return
    }

    // Получаем объект текущего пользователя
    var follower models.User
    if err := db.First(&follower, followerID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Current user not found"})
        return
    }

    // 💥 Отправляем WS уведомление
    wsservice.SendNotification(followeeID, map[string]interface{}{
        "type":      "follow",
        "username":  follower.Username,
        "timestamp": time.Now().Unix(),
    })

    c.JSON(http.StatusOK, gin.H{"message": "Followed successfully"})
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