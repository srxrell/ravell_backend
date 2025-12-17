package handlers

import (
	"go_stories_api/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func UpdateStreak(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    userID := c.MustGet("user_id").(uint)

    var profile models.Profile
    if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile not found"})
        return
    }

    now := time.Now().UTC()
    last := profile.LastActiveAt

    oneDay := 24 * time.Hour
    rewardGiven := false

    if last.IsZero() || now.Sub(last) >= oneDay {
        profile.StreakCount += 1
        profile.LastActiveAt = now

        if profile.StreakCount >= 7 && !profile.StreakRewarded {
            profile.StreakRewarded = true
            rewardGiven = true
        }
    }

    if err := db.Save(&profile).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update streak"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "streak_count": profile.StreakCount,
        "last_active":  profile.LastActiveAt,
        "rewarded":     rewardGiven,
    })
}

func GetUserStreak(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var profile models.Profile
	if err := db.Where("user_id = ?", uint(userID)).First(&profile).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"streak_count": 0})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"streak_count": profile.StreakCount,
		"last_active":  profile.LastActiveAt,
		"rewarded":     profile.StreakRewarded,
	})
}


func GetStreak(c *gin.Context) {
    db := c.MustGet("db").(*gorm.DB)
    userID := c.MustGet("user_id").(uint)

    var profile models.Profile
    if err := db.Where("user_id = ?", userID).First(&profile).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Profile not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "streak_count": profile.StreakCount,
        "last_active":  profile.LastActiveAt,
        "rewarded":     profile.StreakRewarded,
    })
}
