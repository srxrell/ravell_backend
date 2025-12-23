package handlers

import (
	"go_stories_api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Возвращаем ачивки пользователя, создавая их на лету если их нет
func GetUserAchievementsByID(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	idParam := c.Param("id")

	userID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	var userAchievements []models.UserAchievement
	err = db.Preload("Achievement").Where("user_id = ?", userID).Find(&userAchievements).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	// Если нет записей — создаем на лету
	if len(userAchievements) == 0 {
		var allAchievements []models.Achievement
		db.Find(&allAchievements)

		userAchievements = make([]models.UserAchievement, len(allAchievements))
		for i, a := range allAchievements {
			userAchievements[i] = models.UserAchievement{
				UserID:        uint(userID),
				AchievementID: a.ID,
				Progress:      0,
				Unlocked:      false,
				Achievement:   a,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"achievements": userAchievements})
}
func UpdateAllAchievements(db *gorm.DB) {
    var achievements []Achievement
    db.Find(&achievements)

    var users []User
    db.Find(&users)

    for _, ach := range achievements {
        for _, user := range users {
            progress := calculateProgress(db, user.ID, ach)
            var ua UserAchievement
            err := db.Where("user_id = ? AND achievement_id = ?", user.ID, ach.ID).First(&ua).Error
            if err == gorm.ErrRecordNotFound {
                ua = UserAchievement{
                    UserID: user.ID,
                    AchievementID: ach.ID,
                    Progress: progress,
                    Unlocked: progress >= 1,
                }
                db.Create(&ua)
            } else {
                ua.Progress = progress
                if progress >= 1 {
                    ua.Unlocked = true
                    ua.Progress = 1
                }
                db.Save(&ua)
            }
        }
    }
}

// Пример функции calculateProgress
func calculateProgress(db *gorm.DB, userID uint, ach Achievement) float64 {
    var cond map[string]interface{}
    json.Unmarshal(ach.Condition, &cond)

    switch cond["type"] {
    case "story_count":
        var count int64
        db.Model(&Story{}).Where("user_id = ?", userID).Count(&count)
        target := int64(cond["value"].(float64))
        return float64(count)/float64(target)
    }
    return 0
}

// Обновление прогресса ачивки
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
