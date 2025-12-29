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

// Получение ачивок пользователя
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

	// Создаем ачивки на лету, если их нет
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

// Обновление прогресса всех ачивок для всех пользователей
func UpdateAllAchievements(db *gorm.DB) {
	var achievements []models.Achievement

	var users []models.User

	for _, ach := range achievements {
		for _, user := range users {
			progress := calculateProgress(db, user.ID, ach)
			UpdateAchievementProgress(db, user.ID, ach.Key, progress)
		}
	}
}

// Обновление прогресса одной ачивки конкретного пользователя
func UpdateAchievementProgress(db *gorm.DB, userID uint, key string, progress float64) {
	var ach models.Achievement
	if err := db.Where("key = ?", key).First(&ach).Error; err != nil {
		return
	}

	var userAch models.UserAchievement
	err := db.Where("user_id = ? AND achievement_id = ?", userID, ach.ID).First(&userAch).Error
	if err != nil {
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
		Icon:        input.Icon,
		Condition:   datatypes.JSON(input.Condition),
	}

	if err := db.Create(&ach).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create achievement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"achievement": ach})
}
func AddAchievementToUser(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)

	var input struct {
		UserID       uint   `json:"user_id" binding:"required"`
		AchievementKey string `json:"key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, существует ли ачивка
	var ach models.Achievement
	if err := db.Where("key = ?", input.AchievementKey).First(&ach).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Achievement not found"})
		return
	}

	// Проверяем, существует ли уже UserAchievement
	var userAch models.UserAchievement
	err := db.Where("user_id = ? AND achievement_id = ?", input.UserID, ach.ID).First(&userAch).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already has this achievement"})
		return
	}

	if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	// Создаём UserAchievement
	userAch = models.UserAchievement{
		UserID:        input.UserID,
		AchievementID: ach.ID,
		Progress:      0,
		Unlocked:      false,
	}

	if err := db.Create(&userAch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Cannot create UserAchievement"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Achievement added to user", "user_achievement": userAch})
}
// Подсчет прогресса ачивки
func calculateProgress(db *gorm.DB, userID uint, ach models.Achievement) float64 {
	var cond map[string]interface{}
	if err := json.Unmarshal(ach.Condition, &cond); err != nil {
		return 0
	}

	switch cond["type"] {
	case "story_count":
		var count int64
		db.Model(&models.Story{}).Where("user_id = ?", userID).Count(&count)
		target, ok := cond["value"].(float64)
		if !ok || target == 0 {
			return 0
		}
		progress := float64(count) / target
		if progress > 1 {
			progress = 1
		}
		return progress
	}

	return 0
}
