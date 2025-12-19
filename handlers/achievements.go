package handlers

import (
	"fmt"
	"go_stories_api/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func GetUserAchievementsByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idParam := c.Param("id")

	var userID uint
	if _, err := fmt.Sscan(idParam, &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var userAchievements []models.UserAchievement
	db.Preload("Achievement").Where("user_id = ?", userID).Find(&userAchievements)

	c.JSON(http.StatusOK, gin.H{"achievements": userAchievements})
}


func UpdateAchievementProgress(db *gorm.DB, userID uint, key string, progress float64) {
	var ach models.Achievement
	if err := db.Where("key = ?", key).First(&ach).Error; err != nil {
		return
	}

	var userAch models.UserAchievement
	if err := db.Where("user_id = ? AND achievement_id = ?", userID, ach.ID).First(&userAch).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			userAch = models.UserAchievement{
				UserID:        userID,
				AchievementID: ach.ID,
				Progress:      progress,
				Unlocked:      progress >= 1,
			}
			db.Create(&userAch)
			return
		}
	}

	userAch.Progress = progress
	if progress >= 1 {
		userAch.Unlocked = true
		userAch.Progress = 1
	}
	db.Save(&userAch)
}
