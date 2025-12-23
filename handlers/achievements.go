package handlers

import (
	"encoding/json"
	"go_stories_api/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
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

	if len(userAchievements) == 0 {
		var allAchievements []models.Achievement
		db.Find(&allAchievements)

		for _, a := range allAchievements {
			userAchievements = append(userAchievements, models.UserAchievement{
				UserID:        uint(userID),
				AchievementID: a.ID,
				Progress:      0,
				Unlocked:      false,
				Achievement:   a,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"achievements": userAchievements})
}

// Обновление прогресса всех ачивок для всех пользователей
func UpdateAllAchievements(db *gorm.DB) {
	var achievements []models.Achievement
	db.Find(&achievements)

	var users []models.User
	db.Find(&users)

	for _, ach := range achievements {
		for _, user := range users {
			UpdateAchievementProgress(db, user.ID, ach.ID)
		}
	}
}

// Обновление прогресса одной ачивки для конкретного пользователя
func UpdateAchievementProgress(db *gorm.DB, userID uint, achievementID uint) {
	var ach models.Achievement
	if err := db.First(&ach, achievementID).Error; err != nil {
		return
	}

	progress := calculateProgress(db, userID, ach)

	var ua models.UserAchievement
	err := db.Where("user_id = ? AND achievement_id = ?", userID, achievementID).First(&ua).Error
	if err == gorm.ErrRecordNotFound {
		ua = models.UserAchievement{
			UserID:        userID,
			AchievementID: achievementID,
			Progress:      progress,
			Unlocked:      progress >= 1,
		}
		db.Create(&ua)
	} else {
		// обновляем прогресс и unlocked, если уже есть запись
		ua.Progress = progress
		if progress >= 1 {
			ua.Unlocked = true
		}
		db.Save(&ua)
	}
}

// Создание ачивки через API
func CreateAchievement(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		Key         string          `json:"key" binding:"required"`
		Title       string          `json:"title" binding:"required"`
		Description string          `json:"description"`
		Icon        string          `json:"icon"`
		Condition   json.RawMessage `json:"condition"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var exist models.Achievement
	if err := db.Where("key = ?", input.Key).First(&exist).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "achievement already exists"})
		return
	}

	ach := models.Achievement{
		Key:         input.Key,
		Title:       input.Title,
		Description: input.Description,
		IconURL:     input.Icon,
		Condition:   datatypes.JSON(input.Condition),
	}

	if err := db.Create(&ach).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create achievement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"achievement": ach})
}

// Пример функции calculateProgress
func calculateProgress(db *gorm.DB, userID uint, ach models.Achievement) float64 {
	var cond map[string]interface{}
	if err := json.Unmarshal(ach.Condition, &cond); err != nil {
		return 0
	}

	switch cond["type"] {
	case "story_count":
		var count int64
		db.Model(&models.Story{}).Where("user_id = ?", userID).Count(&count)
		target := int64(cond["value"].(float64))
		if target == 0 {
			return 0
		}
		progress := float64(count) / float64(target)
		if progress > 1 {
			progress = 1
		}
		return progress
	}

	return 0
}
