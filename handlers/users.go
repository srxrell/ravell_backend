package handlers

import (
	"go_stories_api/models"
	"net/http"
	"log"
	"net/smtp"
	"os"
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

func send(body string, to string) {
    if to == "" {
        log.Println("Email is empty, skipping sending")
        return
    }

    from := "serrelvorne@gmail.com"
    pass := os.Getenv("EMAIL_PASS") // Убедитесь, что здесь App Password

    msg := "From: " + from + "\n" +
        "To: " + to + "\n" +
        "Subject: Notification from Ravell\n\n" +
        body

    err := smtp.SendMail("smtp.gmail.com:587",
        smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
        from, []string{to}, []byte(msg))

    if err != nil {
        log.Printf("smtp error: %s", err)
        return
    }
    log.Println("Successfully sent email to " + to)
}

func FollowUser(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    followerID := c.MustGet("user_id").(uint)

    // 1. Сначала получаем ТЕКУЩЕГО пользователя (Кто подписывается)
    var follower models.User
    if err := db.First(&follower, followerID).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Current user not found"})
        return
    }

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

    // 2. Получаем ТОГО, НА КОГО подписываемся (нужен Email)
    var followee models.User
    if err := db.First(&followee, followeeID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User to follow not found"})
        return
    }

    // Проверка существующей подписки
    var existingSub models.Subscription
    if err := db.Where("follower_id = ? AND following_id = ?", followerID, followeeID).
        First(&existingSub).Error; err == nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Already following this user"})
        return
    }

    // Создаем подписку
    subscription := models.Subscription{
        FollowerID:  followerID,
        FollowingID: followeeID,
    }

    if err := db.Create(&subscription).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to follow"})
        return
    }

    // 3. Теперь переменные follower и followee существуют, можно отправлять
    // Запускаем в горутине, чтобы клиент не ждал отправки письма
    go send("User @"+follower.Username+" followed you!", followee.Email)

    // Логика для Push уведомлений...
    var devices []models.UserDevice
    db.Where("user_id = ?", followeeID).Find(&devices)
    // ... тут ваш код отправки пушей ...

    c.JSON(http.StatusOK, gin.H{"message": "Followed successfully"})
}

func ActivateInfluencer(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	type Body struct {
		Username    string `json:"username" binding:"required"`
		Title       string `json:"title" binding:"required"`
		Description string `json:"description"`
	}

	var body Body
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid body"})
		return
	}

	var user models.User
	if err := db.Preload("Profile").
		Where("username = ?", body.Username).
		First(&user).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// 1️⃣ Feature (upsert)
	var feature models.Feature
	err := db.Where("user_id = ?", user.ID).First(&feature).Error

	if err == gorm.ErrRecordNotFound {
		feature = models.Feature{
			UserID:        user.ID,
			Title:         body.Title,
			Description:   body.Description,
			UsedInRelease: true,
		}
		db.Create(&feature)
	} else {
		db.Model(&feature).Updates(map[string]any{
			"title":            body.Title,
			"description":      body.Description,
			"used_in_release":  true,
		})
	}

	// 2️⃣ Profile → early
	db.Model(&models.Profile{}).
		Where("user_id = ?", user.ID).
		Update("is_early", true)

	c.JSON(http.StatusOK, gin.H{
		"message":  "Influencer activated",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func UnfollowUser(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)

    currentUserID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
        return
    }
    
    targetUserID := c.Param("id") // Это string

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

    // --- ИСПРАВЛЕНИЕ: Чтобы отправить письмо, нужно найти данные пользователей ---
    
    // 1. Ищем текущего юзера (follower), чтобы узнать его Username
    var currentUser models.User
    if err := db.First(&currentUser, currentUserID).Error; err == nil {
        
        // 2. Ищем того, от кого отписались (followee), чтобы узнать его Email
        var targetUser models.User
        if err := db.First(&targetUser, targetUserID).Error; err == nil {
            
            // 3. Отправляем письмо
            go send("User @"+currentUser.Username+" unfollowed you.", targetUser.Email)
        }
    }
    // ---------------------------------------------------------------------------

    c.JSON(http.StatusOK, gin.H{
        "message": "Successfully unfollowed user",
    })
}

func GetActiveInfluencers(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)

    var users []models.User
db.Preload("Profile").
   Preload("Feature").
   Joins("JOIN features ON features.user_id = users.id").
   Joins("JOIN profiles ON profiles.user_id = users.id").
   Where("features.used_in_release = ?", true).
   Where("profiles.is_early = ?", true).
   Group("users.id").
   Limit(20).
   Find(&users)

    result := make([]gin.H, 0)

    for _, u := range users {
        var storyCount int64
        db.Model(&models.Story{}).Where("user_id = ?", u.ID).Count(&storyCount)

        currentUserID, _ := c.Get("user_id")
        var isFollowing bool
        if currentUserID != nil {
            var exists int64
            db.Model(&models.Subscription{}).
                Where("follower_id = ? AND following_id = ?", currentUserID, u.ID).
                Count(&exists)
            isFollowing = exists > 0
        }

        result = append(result, gin.H{
    "id":           u.ID,
    "username":     u.Username,
    "avatar":       u.Profile.Avatar,
    "story_count":  storyCount,
    "is_following": isFollowing,
    "is_early":     u.Profile.IsEarly,
    "title":        u.Feature.Title,
    "description":  u.Feature.Description,
})
    }

    c.JSON(http.StatusOK, gin.H{
        "influencers": result, // всегда []
    })
}

func AddInfluencer(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)

    type requestBody struct {
        Username    string `json:"username" binding:"required"`
        Title       string `json:"title" binding:"required"`
        Description string `json:"description"`
    }

    var body requestBody
    if err := c.ShouldBindJSON(&body); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }

    var user models.User
    if err := db.Where("username = ?", body.Username).First(&user).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
        return
    }

    feature := models.Feature{
        UserID:        user.ID,
        Title:         body.Title,
        Description:   body.Description,
        UsedInRelease: false,
        CreatedAt:     time.Now(),
        UpdatedAt:     time.Now(),
    }

    if err := db.Create(&feature).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Influencer added successfully",
        "feature": feature,
    })
}